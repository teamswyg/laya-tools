"""Policy contract test with a synthetic module, not a Laya integration test."""
import unittest
import torch
from head_policy import configure_head_only


class HeadPolicyTest(unittest.TestCase):
    def test_freezes_encoder_final_norm_and_action_head(self):
        model = torch.nn.Module()
        model.encoder = torch.nn.Sequential(torch.nn.Linear(4, 4), torch.nn.LayerNorm(4))
        model.head = torch.nn.Linear(4, 4)
        model.type_emb = torch.nn.Embedding(3, 4)
        model.scorer = torch.nn.Linear(4, 1)
        model.act_head = torch.nn.Linear(4, 2)
        names = configure_head_only(model)
        self.assertTrue(names)
        self.assertTrue(all(n.startswith(('head.', 'type_emb.', 'scorer.')) for n in names))
        self.assertFalse(model.encoder.training)
        self.assertTrue(model.head.training)
        self.assertTrue(all(not p.requires_grad for p in model.encoder.parameters()))
        self.assertTrue(all(not p.requires_grad for p in model.act_head.parameters()))
        before = {n: p.detach().clone() for n, p in model.named_parameters()}
        optimizer = torch.optim.AdamW((p for p in model.parameters() if p.requires_grad), lr=0.01)
        features = model.encoder(torch.randn(8, 2, 4))
        features = features + model.type_emb(torch.zeros(8, dtype=torch.long))[:, None, :]
        logits = model.scorer(model.head(features)).squeeze(-1)
        loss = torch.nn.functional.cross_entropy(logits, torch.arange(8) % 2)
        loss.backward()
        optimizer.step()
        for prefix in ('head.', 'type_emb.', 'scorer.'):
            self.assertTrue(any(not torch.equal(before[n], p.detach())
                                for n, p in model.named_parameters() if n.startswith(prefix)))
        for n, p in model.named_parameters():
            if not p.requires_grad:
                self.assertTrue(torch.equal(before[n], p.detach()), n)
                self.assertIsNone(p.grad)

    def test_unknown_structure_fails(self):
        with self.assertRaises(ValueError):
            configure_head_only(torch.nn.Linear(4, 4))


if __name__ == '__main__':
    unittest.main()
