# Nano Banana image prompts — `/product/what-are-feature-flags`

Two illustrations are still placeholders on the page. Generate them with Nano Banana
(Google Gemini image generation), then drop the file into this folder and wire it in by
replacing the placeholder's `placeholderLabel` with an `imageSrc`:

```js
import deployVsRelease from './deploy-vs-release.png';
// ...
<FeatureRow ... imageSrc={deployVsRelease} imageAlt="..." />
```

All prompts share the existing house style (the honeycomb `use-cases.png`): flat vector,
**dual-theme palette** — medium-saturation teal, purple, slate-blue, and coral accent;
mid-tone flat solid fills only (no pastels, no near-white, no near-black) so shapes read
clearly after chromakey on both light and dark page backgrounds; crisp icons,
**no text or letters**, plenty of whitespace,
**solid #00FF00 chromakey green background with zero gradients, zero lighting effects,
and zero shadows**, 4:3.

---

## Prompt A — Decouple deploy from release (section "Decouple deploy from release")

> Flat-vector illustration: decouple deployment from release, with kill-switch as the
> hero capability. Left — a shipped-code icon (package or deploy rocket) already sitting
> in production. Horizontal arrow to a large central feature-flag toggle acting as a gate
> between deploy and release. From the toggle, two clear storylines: (1) upper path —
> progressive release with three branching paths to increasingly larger user-group dot
> clusters (small team → small cohort → everyone), all flowing forward while the toggle
> stays ON; (2) lower path — kill switch: the same toggle flipped OFF, a bold severed
> connection line cutting flow from a misbehaving feature tile (small warning spark or
> glitch motif) to user dots, with a power-off icon beside the toggle to convey instant
> shutdown in production without redeploy or rollback. Medium-saturation dual-theme palette:
> medium teal, medium purple (#6c46f2), medium slate-blue, coral accent for the kill-switch
> path; mid-tone flat solid fills only (no pastels, no near-white, no near-black);
> connection lines in medium slate; simple line + filled icons; no photorealism;
> no human faces; no text or letters; balanced composition;
> solid #00FF00 chromakey green background with zero gradients, zero lighting effects,
> and zero shadows; aspect ratio 4:3.

Save as: `deploy-vs-release.png`

---

## Prompt B — Self-hosted architecture (section "Feature flags with GO Feature Flag")

> Flat-vector technical diagram of a self-hosted feature-flag architecture. Center — one
> server/container tile (relay proxy) reading from a YAML config file icon on the left
> (file with small toggle switches). Connection lines fan out right to a row of application
> tiles (generic code and curly-brace icons, a few stacked). Convey: no database,
> file-based, runs on your own infrastructure. No cloud-vendor logos. Medium-saturation
> dual-theme palette: medium teal, medium purple (#6c46f2), medium slate-blue, coral accent
> for highlights; mid-tone flat solid fills only (no pastels, no near-white, no near-black);
> connection lines in medium slate; crisp vector icons; no photorealism;
> no text or letters; plenty of whitespace;
> solid #00FF00 chromakey green background with zero gradients, zero lighting effects,
> and zero shadows; aspect ratio 4:3.

Save as: `self-hosted-architecture.png`

---

## Prompt C — (optional) Evaluation flow

Only needed if the "How feature flags work" section is later turned into a text+illustration split.

> Flat-vector diagram of feature-flag evaluation at runtime. Left — evaluation-context card
> icon (user/id tag with small attribute chips). Arrow to a central decision node with
> branching targeting rules (flowchart fork). Branches lead to outcome tiles (on/off toggle
> and value chips). Medium-saturation dual-theme palette: medium teal, medium purple
> (#6c46f2), medium slate-blue, coral accent for decision highlights; mid-tone flat solid
> fills only (no pastels, no near-white, no near-black); connection lines in medium slate;
> simple vector icons; no photorealism; no text or letters; balanced layout;
> solid #00FF00 chromakey green background with zero gradients, zero lighting effects,
> and zero shadows; aspect ratio 4:3.

Save as: `evaluation-flow.png`

---

### Tips
- Keep "no text or letters" in the prompt — Nano Banana renders text unreliably; any labels
  are better added in HTML later.
- Repeat the chromakey line verbatim — models often add gradients, shadows, or lighting
  unless told not to.
- Repeat the dual-theme color constraints — models default to light pastels or dark navy that
  vanish on one theme; mid-tone teal / purple / slate-blue / coral reads on both.
- Key out `#00FF00` in post (or export with transparency) before placing on the site.
- After generating, downscale to ~1000px wide and recompress (e.g. `sips --resampleWidth 1000`)
  to keep the asset small.
