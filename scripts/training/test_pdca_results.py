"""Audit saved public measurements without loading weights or installing torch."""
import hashlib
import json
from pathlib import Path
import unittest

ROOT = Path(__file__).resolve().parents[2]


class MeasurementAudit(unittest.TestCase):
    def test_saved_metrics_match_predictions_and_frozen_selection(self):
        reports = sorted((ROOT / 'benchmarks/results').glob('pdca-*/seed*-evaluation.json'))
        self.assertTrue(reports, 'Expected archived experiments')
        for path in reports:
            with self.subTest(report=str(path.relative_to(ROOT))):
                r = json.loads(path.read_text())
                selection = json.loads(path.with_name(path.name.replace('evaluation', 'selection')).read_text())
                data = ROOT / 'benchmarks/training' / path.parent.name
                for key, name in [('data_sha256', 'families.json'), ('plan_sha256', 'plan.json')]:
                    self.assertEqual(r[key], hashlib.sha256((data/name).read_bytes()).hexdigest())
                    self.assertEqual(r[key], selection[key])
                self.assertEqual(r['selected_head_sha256'], selection['head_sha256'])
                self.assertFalse(selection['final_test_used_for_selection'])
                self.assertEqual(r['passed'], all(r['criteria'].values()))
                self.assertAlmostEqual(r['permutation_flip_rate'],
                                       r['permutation_flips']/r['permutation_comparisons'])
                for key in ['baseline', 'tuned', 'base_legacy', 'tuned_legacy']:
                    m = r[key]
                    predictions = m['predictions']
                    self.assertEqual(m['n'], len(predictions))
                    correct = sum(p['expected'] == p['predicted'] for p in predictions)
                    accepted = [p for p in predictions if max(p['probabilities']) >= .9]
                    ac = sum(p['expected'] == p['predicted'] for p in accepted)
                    self.assertEqual(m['correct'], correct)
                    self.assertEqual(m['accepted'], len(accepted))
                    self.assertEqual(m['accepted_correct'], ac)
                    self.assertAlmostEqual(m['accuracy'], correct/m['n'], places=6)
                    self.assertAlmostEqual(m['coverage'], len(accepted)/m['n'], places=6)
                    if accepted:
                        self.assertAlmostEqual(m['accepted_precision'], ac/len(accepted), places=6)
                    else:
                        self.assertIsNone(m['accepted_precision'])
                    for p in predictions:
                        self.assertEqual(p['predicted'], max(range(3), key=p['probabilities'].__getitem__))
                        self.assertAlmostEqual(sum(p['probabilities']), 1, places=6)


if __name__ == '__main__':
    unittest.main()
