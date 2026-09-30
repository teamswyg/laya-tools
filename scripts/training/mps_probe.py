"""Maintainer-only MPS readiness probe. This does NOT load or train Laya."""
import argparse
import json
import math
import os
import platform
import resource
import time


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--steps', type=int, default=3, choices=range(1, 11))
    args = parser.parse_args()
    if os.environ.get('PYTORCH_ENABLE_MPS_FALLBACK', '0') != '0':
        parser.error('CPU fallback must be disabled for an MPS readiness claim')
    os.environ['PYTORCH_ENABLE_MPS_FALLBACK'] = '0'
    import torch
    if platform.system() != 'Darwin' or not torch.backends.mps.is_available():
        parser.error('This probe requires available Apple MPS; no CPU substitution')
    torch.manual_seed(1729)
    torch.mps.set_per_process_memory_fraction(0.25)
    device = torch.device('mps')
    encoder = torch.nn.Linear(64, 128).float().to(device)
    head = torch.nn.Linear(128, 4).float().to(device)
    encoder.requires_grad_(False)
    frozen = {n: p.detach().cpu().clone() for n, p in encoder.named_parameters()}
    before = head.weight.detach().cpu().clone()
    optimizer = torch.optim.AdamW(head.parameters(), lr=1e-3)
    x = torch.randn(8, 64, device=device)
    y = torch.arange(8, device=device) % 4
    losses = []
    allocated = []
    driver = []
    torch.mps.synchronize()
    started = time.perf_counter()
    for _ in range(args.steps):
        optimizer.zero_grad(set_to_none=True)
        logits = head(encoder(x))
        loss = torch.nn.functional.cross_entropy(logits, y)
        loss.backward()
        if not all(p.grad is not None and torch.isfinite(p.grad).all().item()
                   for p in head.parameters()):
            raise RuntimeError('Missing or non-finite head gradients')
        optimizer.step()
        torch.mps.synchronize()
        losses.append(float(loss.item()))
        allocated.append(torch.mps.current_allocated_memory())
        driver.append(torch.mps.driver_allocated_memory())
    changed = not torch.equal(before, head.weight.detach().cpu())
    unchanged = all(torch.equal(frozen[n], p.detach().cpu()) and p.grad is None
                    for n, p in encoder.named_parameters())
    if not changed or not unchanged or not all(math.isfinite(v) for v in losses):
        raise RuntimeError('Weight update/freeze/finite-loss verification failed')
    print(json.dumps({
        'schema_version': 1, 'probe': 'synthetic_mps_backward',
        'laya_loaded': False, 'laya_training_verified': False,
        'device': str(device), 'cpu_fallback': False,
        'python': platform.python_version(), 'torch': torch.__version__,
        'os': platform.system(), 'os_version': platform.mac_ver()[0],
        'architecture': platform.machine(), 'dtype': 'float32',
        'steps': args.steps, 'losses': losses, 'head_weights_changed': changed,
        'frozen_encoder_unchanged': unchanged,
        'synchronized_probe_loop_seconds': time.perf_counter() - started,
        'process_peak_rss_bytes': resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,
        'sampled_mps_allocated_max_bytes': max(allocated),
        'sampled_mps_driver_allocated_max_bytes': max(driver),
        'memory_note': 'Step-end samples are not GPU peaks. Unified RSS/MPS memory overlaps; do not add.',
    }, indent=2))


if __name__ == '__main__':
    main()
