"""Offline checks for fixed PDCA boundaries; these do not validate semantic labels."""
import hashlib
import json
from pathlib import Path
import unittest

ROOT = Path(__file__).resolve().parents[2]


class PDCAContractTest(unittest.TestCase):
    def test_preregistered_files_and_group_boundaries(self):
        root = ROOT / 'benchmarks/training/pdca-01'
        expected = {'plan.json': '5a5c019a57e19718c0ff25fe04027329206a6503627fda5b1fcf81d0c8a352ed',
                    'families.json': 'f907a7088b75e7509f7ea3b7b2827d45e5cbdb1b87591cf7dc73fbd24715f3b7'}
        for name, sha in expected.items():
            self.assertEqual(hashlib.sha256((root/name).read_bytes()).hexdigest(), sha)
        families = json.loads((root/'families.json').read_text())
        self.assertEqual(len(families), 28)
        self.assertEqual(len({f['family'] for f in families}), 28)
        cases = [text for f in families for text in f['cases']]
        self.assertEqual(len(cases), len(set(cases)))
        self.assertEqual(len(cases), 84)
        self.assertEqual({s:sum(len(f['cases']) for f in families if f['split']==s)
                          for s in ('train','validation','calibration','test')},
                         {'train':36,'validation':12,'calibration':12,'test':24})

    def test_fixed_acceptance_includes_safety_and_coverage(self):
        p = json.loads((ROOT/'benchmarks/training/pdca-01/plan.json').read_text())
        self.assertEqual(p['gate'], .9)
        self.assertEqual(p['seeds'], [1729,2718])
        self.assertTrue(p['criteria']['both_seeds_must_pass'])
        self.assertGreater(p['criteria']['minimum_coverage'], 0)
        self.assertEqual(p['criteria']['strong_to_fast_errors'], 0)

    def test_second_cycle_preserves_gates_and_moves_exposed_test(self):
        first = ROOT / 'benchmarks/training/pdca-01'
        second = ROOT / 'benchmarks/training/pdca-02'
        for name, sha in {
            'plan.json': 'c57f0f5583335ff260dcf23f026c26dd626249bf6fc2e3937d2b0efbead06eda',
            'families.json': '6308c542bf9eae7c445d7a7279211fe88d6d3845343dee3be987c1339be01d6e',
        }.items():
            self.assertEqual(hashlib.sha256((second/name).read_bytes()).hexdigest(), sha)
        a = json.loads((first/'plan.json').read_text())
        b = json.loads((second/'plan.json').read_text())
        self.assertEqual(a['criteria'], b['criteria'])
        self.assertEqual(a['gate'], b['gate'])
        old = json.loads((first/'families.json').read_text())
        new = json.loads((second/'families.json').read_text())
        self.assertEqual(len(new), len({f['family'] for f in new}))
        previous_test = {f['family'] for f in old if f['split']=='test'}
        self.assertTrue(previous_test <= {f['family'] for f in new if f['split']=='validation'})
        self.assertFalse({f['family'] for f in old} & {f['family'] for f in new if f['split']=='test'})
        self.assertEqual(sum(len(f['cases']) for f in new if f['split']=='test'), 24)

    def test_third_cycle_keeps_gates_and_new_final(self):
        root = ROOT / 'benchmarks/training/pdca-03'
        for name, sha in {
            'plan.json': '1e8cd9ae659d263af060c99ea4860afebf7fe4bc1ae5fb4243193db1ba74ccd1',
            'families.json': '1fdb16ac08de923431c7262f6b0bf7dd7a51feb7b5dc1d13ecab7bcda78b005c',
        }.items():
            self.assertEqual(hashlib.sha256((root/name).read_bytes()).hexdigest(), sha)
        plan = json.loads((root/'plan.json').read_text())
        first = json.loads((ROOT/'benchmarks/training/pdca-01/plan.json').read_text())
        self.assertEqual(plan['criteria'], first['criteria'])
        self.assertEqual(plan['gate'], first['gate'])
        families = json.loads((root/'families.json').read_text())
        old = json.loads((ROOT/'benchmarks/training/pdca-02/families.json').read_text())
        self.assertFalse({f['family'] for f in old} & {f['family'] for f in families if f['split']=='test'})
        self.assertEqual(len(families), len({f['family'] for f in families}))
        texts = [t for f in families for t in f['cases']]
        self.assertEqual(len(texts), len(set(texts)))
        self.assertEqual({s:sum(len(f['cases']) for f in families if f['split']==s)
                          for s in ('train','validation','calibration','test')},
                         {'train':108,'validation':24,'calibration':12,'test':24})


if __name__=='__main__': unittest.main()
