import itertools
from types import SimpleNamespace
import unittest
import torch
from fixed_tier import FixedTier, state_feature


class Tokenizer:
    cls_token_id, sep_token_id = 101, 102
    def __init__(self, ids): self.ids=ids
    def __call__(self,*args,**kwargs):
        ids=torch.tensor([self.ids])
        return dict(input_ids=ids,attention_mask=torch.ones_like(ids))


class Encoder(torch.nn.Module):
    def __init__(self):
        super().__init__(); self.anchor=torch.nn.Parameter(torch.zeros(1))
    def forward(self,input_ids,attention_mask):
        return SimpleNamespace(last_hidden_state=input_ids[...,None].float().expand(-1,-1,2))


class Model(torch.nn.Module):
    def __init__(self):
        super().__init__();self.encoder=Encoder()


class FixedTierTest(unittest.TestCase):
    def test_pool_excludes_special_tokens_and_rejects_empty_or_truncated(self):
        m=Model()
        self.assertTrue(torch.equal(state_feature(m,Tokenizer([101,7,8,102]),'text'),torch.tensor([7.5,7.5])))
        for ids in ([101,102],[7]*513):
            with self.assertRaises(ValueError):state_feature(m,Tokenizer(ids),'text')

    def test_fixed_output_dimensions_and_serialization_permutations(self):
        h=FixedTier()
        self.assertEqual(sum(p.numel() for p in h.parameters()),5123)
        z=h(torch.randn(4,1024))
        for order in itertools.permutations(range(3)):
            back=z[:,list(order)][:,[order.index(i) for i in range(3)]]
            self.assertTrue(torch.equal(back,z))


if __name__=='__main__':unittest.main()
