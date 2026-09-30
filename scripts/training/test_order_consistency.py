"""Meaningful invariants of the optional differentiable consistency objective."""
import unittest
import torch
from pdca_train import order_js


class ConsistencyTest(unittest.TestCase):
    def test_matching_orders_zero_and_conflicts_penalized(self):
        matching = torch.tensor([[4., 0., -1.]] * 3)
        self.assertAlmostEqual(float(order_js(matching)), 0, places=6)
        conflict = torch.tensor([[4., 0., -1.], [-1., 4., 0.], [0., -1., 4.]], requires_grad=True)
        loss = order_js(conflict)
        self.assertGreater(float(loss.detach()), .5)
        loss.backward()
        self.assertTrue(torch.isfinite(conflict.grad).all())
        self.assertTrue((conflict.grad.abs().sum(-1) > 0).all())

    def test_shared_logit_shift_and_order_relabeling_do_not_change_penalty(self):
        z = torch.tensor([[[4., 0., -1.], [1., 2., 3.]], [[-1., 4., 0.], [2., 1., 3.]]])
        self.assertAlmostEqual(float(order_js(z)), float(order_js(z + 10)), places=6)
        self.assertAlmostEqual(float(order_js(z)), float(order_js(z.flip(0).flip(-1))), places=6)


if __name__ == '__main__':
    unittest.main()
