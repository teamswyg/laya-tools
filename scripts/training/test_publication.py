"""Offline standard-library publication and data-boundary tests; no model download."""
import csv
import hashlib
import json
from pathlib import Path
import tempfile
import unittest
from publish_pilot import ALLOWED, validate


class PublicationTest(unittest.TestCase):
    def package(self, root):
        for n in ALLOWED - {'MANIFEST.json'}:
            (root / n).write_text('public fixture\n')
        data_hash = hashlib.sha256((root / 'training-data.tsv').read_bytes()).hexdigest()
        head_hash = hashlib.sha256((root / 'head.safetensors').read_bytes()).hexdigest()
        (root / 'results.json').write_text(json.dumps({'frozen_weights_unchanged': True, 'head_weights_changed': True,
                                                       'reload_max_logit_error': 0, 'data_sha256': data_hash}))
        (root / 'PROVENANCE.json').write_text(json.dumps({'license': 'apache-2.0', 'dataset_origin': 'original_synthetic',
                                                        'production_ready': False, 'head_sha256': head_hash, 'data_sha256': data_hash}))
        self.manifest(root)

    def manifest(self, root):
        (root / 'MANIFEST.json').write_text(json.dumps({n: hashlib.sha256((root / n).read_bytes()).hexdigest()
                                                      for n in ALLOWED - {'MANIFEST.json'}}))

    def test_tampered_weight_rejected(self):
        with tempfile.TemporaryDirectory() as d:
            root = Path(d); self.package(root)
            validate(root)
            (root / 'head.safetensors').write_text('changed')
            with self.assertRaisesRegex(ValueError, 'hash mismatch'):
                validate(root)

    def test_optimizer_and_symlink_rejected(self):
        with tempfile.TemporaryDirectory() as d:
            root = Path(d); self.package(root)
            (root / 'resume.pt').write_text('not public')
            with self.assertRaises(ValueError): validate(root)
            (root / 'resume.pt').unlink()
            (root / 'NOTICE').unlink()
            (root / 'NOTICE').symlink_to(root / 'LICENSE')
            with self.assertRaises(ValueError): validate(root)

    def test_wrong_dataset_even_with_updated_manifest_rejected(self):
        with tempfile.TemporaryDirectory() as d:
            root = Path(d); self.package(root)
            (root / 'training-data.tsv').write_text('different')
            self.manifest(root)
            with self.assertRaisesRegex(ValueError, 'Dataset provenance'):
                validate(root)

    def test_pilot_data_boundaries(self):
        path = Path(__file__).resolve().parents[2] / 'benchmarks/training/pilot-v1.tsv'
        with path.open() as f: rows = list(csv.DictReader(f, delimiter='\t'))
        self.assertEqual(len(rows), 88)
        self.assertEqual(len({r['state'] for r in rows}), len(rows))
        for task, labels in [('difficulty', {'fast', 'standard', 'strong'}),
                             ('decomposition', {'keep_atomic', 'split_sequential', 'split_parallel', 'needs_context'})]:
            for split in ('train', 'validation', 'calibration', 'test'):
                group = [r for r in rows if r['task'] == task and r['split'] == split]
                self.assertEqual({r['label'] for r in group}, labels)
                self.assertTrue(all(r['state'].isascii() for r in group))


if __name__ == '__main__':
    unittest.main()
