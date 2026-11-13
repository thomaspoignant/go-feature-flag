# Nano-banana image prompts — `/product/ai`

Illustrations for the AI product page (`website/src/pages/product/ai/index.js`).
Each `FeatureRow` currently renders a placeholder box labelled with its prompt letter.
Generate each image, drop the PNG in this folder, then swap the placeholder for an
`imageSrc` import (see "Wiring images in" at the bottom).

## Shared style guide (include in every prompt)

> Clean, modern flat **vector illustration**, soft rounded shapes, subtle long shadows,
> generous negative space. Palette: GO Feature Flag teal/sage — primary `#18b192`,
> deep `#0a5b4f`, sage accent `#9fbeb3`, light neutral background `#f7fbf9`, dark text
> shapes `#273437`. No photorealism, no gradients-heavy 3D, no logos, **no text or
> letters baked into the image**. Centered composition, ~4:3 aspect ratio,
> 1200×896 px, crisp at small sizes.

---

## Prompt A — The core idea (hero feature row)

> A torrent of small code blocks and chat/AI bubbles streaming in from the left like a
> firehose, funneling down through a single large toggle switch in the center, and
> emerging on the right as a calm, orderly single stream reaching a small group of user
> avatars. The switch is the clear focal point. Conveys "AI generates more than you can
> review; one flag controls the flow." Flat vector, teal/sage palette per the style guide.

## Prompt B — AI-generated code (wrap agent output in a flag)

> A friendly robot / coding-agent figure on the left handing off blocks of source code,
> which enter a pipe that is gated by a flag-shaped valve before reaching production
> servers on the right. The valve is mostly closed, letting only a thin controlled
> trickle through to a small "internal team" badge. Conveys "agent-written code merges
> dark behind a flag." Flat vector, teal/sage palette per the style guide.

## Prompt C — Model rollout / canary (roll out to a percentage)

> User traffic (a crowd of small avatars) arriving at a fork that splits into two
> labelled lanes: a wide lane to an existing model chip and a thin lane to a new model
> chip, with a rising ramp/curve showing the thin lane growing over time. A small
> percentage dial sits at the fork. Conveys "canary a new model, then ramp it." Flat
> vector, teal/sage palette per the style guide.

## Prompt D — A/B test models & prompts (experimentation)

> Two AI variants side by side inside a bracketed window — left card shows a chat/prompt
> bubble "A", right card shows "B" — with a balance scale or a small bar chart between
> them comparing results, and clear start and end markers framing the window. Conveys
> "time-boxed experiment measuring which AI wins." Flat vector, teal/sage palette per the
> style guide.

## Prompt E — Kill switch (turn the AI off)

> A large prominent power/kill switch in the foreground being flipped, rerouting flow
> away from a glitching AI/LLM chip (small warning spark) toward a calm, solid
> deterministic "rules engine" gear block that catches the traffic safely. Conveys
> "one flip drops everyone to the safe fallback." Flat vector, teal/sage palette per the
> style guide.

---

## Wiring images in

Once the PNGs exist in `website/static/img/landing/ai/` (e.g. `core-idea.png`,
`ai-code.png`, `model-rollout.png`, `experiment.png`, `kill-switch.png`), in
`index.js`:

1. Add imports at the top, e.g.
   `import coreIdea from '@site/static/img/landing/ai/core-idea.png';`
2. On each matching `FeatureRow`, replace `placeholderLabel="Prompt X - ..."` with:
   `imageSrc={coreIdea} imageAlt="..." imageWidth={1200} imageHeight={896}`
   (keep an honest `imageAlt`; set width/height to the real pixel dimensions to avoid
   layout shift).
