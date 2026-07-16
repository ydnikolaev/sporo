# Recipes — transferable builds (sporo)

This repository authors **recipes** with the sporo CLI: one self-contained document per
capability that teaches an agent in a repository that has never seen this one how to build
the same thing — on its own stack, without repeating this build's mistakes.

When asked to write, check, or hand over a recipe:

- **Read the genre first**: `sporo genre` prints the authoring spec (eleven gated sections;
  the neutrality constraint — the body names roles, never paths/filenames/products; every
  build step ends with `**Done when:**`; every scar is symptom → root cause → fix).
- **Harvest before you recall**: `sporo harvest --since <the release before the work>` mines
  this repo's own record for raw material. Judgment (which failure was structural) is yours;
  memory is not a source.
- **Scaffold, never a blank page**: `sporo new <slug> --from-harvest <file>` writes a coached
  draft with the harvest's scars pre-seeded. A draft (`draft: true`) cannot be sealed or
  exported; removing the mark is how you declare it done.
- **Gate it**: `sporo lint` checks this project's recipes home (default `.sporo/recipes/`)
  for shape, neutrality, and registry coherence.
- **Seal, then export**: `sporo seal <slug>` records version + content hash (a sealed recipe
  never silently mutates); `sporo export <slug>` prints the deliverable — banner stripped,
  adoption protocol appended. Hand over the export, never the source file.
- **Merge what comes back**: `sporo feedback add <slug> <file>` files a reader's report-back;
  its new scars become the recipe's next version (bump `version:`, re-seal).
- **Track what you built from**: `sporo adopt <exported file>` records a handover;
  `sporo pull` checks adopted recipes against their sources — loud when an exact contract
  moved (rebuild, then re-run `sporo conform`).
