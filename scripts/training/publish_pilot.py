"""Validate an explicit experimental package; publish only to a new versioned repo.

No folder-wide upload, optimizer state, credentials or private input is permitted.
"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess

ALLOWED = {'README.md', 'LICENSE', 'NOTICE', 'MODIFICATIONS.md', 'PROVENANCE.json',
           'head.safetensors', 'results.json', 'training-data.tsv', 'requirements.txt',
           'UPSTREAM_MODEL_CARD.md', 'ENCODER_MODEL_CARD.md', 'MANIFEST.json'}


def validate(folder):
    if folder.is_symlink():
        raise ValueError('Package root must not be a symlink')
    entries = list(folder.iterdir())
    if any(p.is_symlink() or not p.is_file() or p.name not in ALLOWED for p in entries):
        raise ValueError('Unexpected package entry')
    if {p.name for p in entries} != ALLOWED:
        raise ValueError('Package file set incomplete')
    manifest = json.loads((folder / 'MANIFEST.json').read_text())
    if set(manifest) != ALLOWED - {'MANIFEST.json'}:
        raise ValueError('Manifest file set mismatch')
    for name, digest in manifest.items():
        with (folder / name).open('rb') as f:
            actual = hashlib.file_digest(f, 'sha256').hexdigest()
        if actual != digest:
            raise ValueError('Package hash mismatch: ' + name)
    result = json.loads((folder / 'results.json').read_text())
    if not result.get('frozen_weights_unchanged') or not result.get('head_weights_changed'):
        raise ValueError('Training integrity checks missing')
    if result.get('reload_max_logit_error', float('inf')) > 1e-4:
        raise ValueError('Checkpoint reload parity missing')
    provenance = json.loads((folder / 'PROVENANCE.json').read_text())
    if provenance.get('license') != 'apache-2.0' or provenance.get('dataset_origin') != 'original_synthetic':
        raise ValueError('License/data review not recorded')
    if provenance.get('production_ready') is not False:
        raise ValueError('Pilot must be labelled non-production')
    if provenance.get('head_sha256') != manifest['head.safetensors']:
        raise ValueError('Head provenance mismatch')
    if provenance.get('data_sha256') != manifest['training-data.tsv'] or result.get('data_sha256') != manifest['training-data.tsv']:
        raise ValueError('Dataset provenance mismatch')
    return manifest


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--package', type=Path, required=True)
    p.add_argument('--repo', required=True)
    p.add_argument('--publish', action='store_true')
    args = p.parse_args()
    if not re.fullmatch(r'JooYoon/riidolaya-(difficulty|decomposition)-pilot-v[0-9]+\.[0-9]+', args.repo):
        raise ValueError('Use a new explicit JooYoon riidolaya pilot version')
    validate(args.package)
    if not args.publish:
        print('Package validated; no publication requested')
        return
    provenance = json.loads((args.package / 'PROVENANCE.json').read_text())
    revision = provenance.get('source_revision', '')
    if not re.fullmatch('[0-9a-f]{40}', revision):
        raise ValueError('An exact source revision with successful CI is required')
    runs = json.loads(subprocess.check_output([
        'gh', 'run', 'list', '--repo', 'teamswyg/laya-tools', '--workflow', 'ci.yml',
        '--commit', revision, '--json', 'status,conclusion', '--limit', '20'], text=True))
    if not any(r['status'] == 'completed' and r['conclusion'] == 'success' for r in runs):
        raise ValueError('Source revision has no successful CI run; publication stopped')
    from huggingface_hub import HfApi
    api = HfApi()
    if api.whoami()['name'] != 'JooYoon':
        raise ValueError('Authenticated identity must be JooYoon')
    if api.repo_exists(args.repo, repo_type='model'):
        raise ValueError('Refusing to overwrite an existing versioned repository')
    api.create_repo(args.repo, repo_type='model', private=False, exist_ok=False)
    commit = api.upload_folder(repo_id=args.repo, repo_type='model', folder_path=str(args.package),
                               allow_patterns=sorted(ALLOWED), commit_message='Publish experimental head-only MPS pilot with provenance and limitations')
    api.create_tag(repo_id=args.repo, tag='pilot-release', revision=commit.oid, repo_type='model')
    print(json.dumps({'repo': args.repo, 'revision': commit.oid, 'url': f'https://huggingface.co/{args.repo}/tree/{commit.oid}'}))


if __name__ == '__main__':
    main()
