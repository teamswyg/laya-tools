"""Head-only trainability policy for the inspected laya 0.3.21 DecisionModel.

No weights are loaded here. Task choice labels train head/type_emb/scorer only;
act_head (answer vs escalation) needs separate targets and remains frozen.
"""


def configure_head_only(model):
    required = ('encoder', 'head', 'type_emb', 'scorer', 'act_head')
    if any(getattr(model, name, None) is None for name in required):
        raise ValueError('Unsupported DecisionModel structure; review the pinned SDK')
    for parameter in model.parameters():
        parameter.requires_grad_(False)
    for name in ('head', 'type_emb', 'scorer'):
        getattr(model, name).requires_grad_(True)
    model.train()
    # Frozen encoder dropout must stay disabled. Call this after model.train().
    model.encoder.eval()
    model.act_head.eval()
    return tuple(name for name, parameter in model.named_parameters()
                 if parameter.requires_grad)
