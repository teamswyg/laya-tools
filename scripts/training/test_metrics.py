"""Maintainer torch tests for the acceptance-boundary arithmetic."""
import unittest
import torch
from train_pilot import metrics


class MetricsTest(unittest.TestCase):
    def test_nine_of_ten_is_exact_threshold_not_float32_underflow(self):
        z=torch.tensor([[20.,0.,0.]]*10)
        r=metrics(z,[0]*9+[1])
        self.assertEqual(r['accepted'],10)
        self.assertEqual(r['accepted_correct'],9)
        self.assertGreaterEqual(r['accepted_precision'], .9)
        self.assertEqual(r['accuracy'], .9)

    def test_no_acceptance_has_no_precision_claim(self):
        r=metrics(torch.zeros((3,3)),[0,1,2])
        self.assertEqual(r['accepted'],0)
        self.assertIsNone(r['accepted_precision'])


if __name__=='__main__': unittest.main()
