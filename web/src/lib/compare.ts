// The positioning SSOT: where a recipe sits among the other ways capability "moves" — a pasted
// prompt, an installed skill, a running MCP server, a fine-tuned model. Three consumers read this
// one table so they can never disagree: the HTML page (compare.astro), its markdown twin
// (compare.md.ts), and the single-fetch agent corpus (llms-full.txt.ts). Cells state MECHANISMS,
// not verdicts — "weights you train and host", not "worse than a recipe" — so the comparison reads
// as fact an agent can act on, not a pitch.

export type CompareRow = { dim: string; cells: string[] }; // cells align to `columns`, in order

// Recipe stays LAST so the page's "own column" highlight (columns[i] === 'Recipe') keeps working.
export const columns = ['Prompt', 'Skill', 'MCP server', 'Fine-tune', 'Recipe'];

export const rows: CompareRow[] = [
  {
    dim: 'What it is',
    cells: [
      'An instruction you paste',
      'A capability installed into your agent',
      'A running service your agent calls',
      'Model weights retrained on your data',
      'A self-contained file that teaches the build',
    ],
  },
  {
    dim: 'Where it runs',
    cells: [
      'Wherever you paste it',
      'The harness it is installed in',
      'Your (or a vendor’s) server',
      'The model you trained and host',
      'Any harness, on the reader’s own stack',
    ],
  },
  {
    dim: 'Crosses to a repo that has never seen yours',
    cells: [
      'No — assumes your context',
      'No — tied to your harness',
      'Only while the server is reachable',
      'No — baked into one model’s weights',
      'Yes — that is the whole point',
    ],
  },
  {
    dim: 'Carries your scars',
    cells: [
      'No',
      'Sometimes, implicitly',
      'No',
      'No — averaged into the weights',
      'Yes — symptom → root cause → fix',
    ],
  },
  {
    dim: 'Versioned & sealed',
    cells: [
      'No',
      'Depends',
      'Depends',
      'By model checkpoint',
      'Yes — id, version + content hash',
    ],
  },
  {
    dim: 'Improves over time',
    cells: [
      'No',
      'By editing in place',
      'By redeploying',
      'By retraining',
      'By report-backs → next version',
    ],
  },
  {
    dim: 'Needs an account / server / runtime',
    cells: ['No', 'No', 'Yes', 'Yes — training + a serving runtime', 'No'],
  },
];

// compareMarkdown renders the table + FAQ as GitHub-flavored markdown — the SSOT form the
// markdown twin (compare.md) serves and the agent corpus (llms-full.txt) folds in, so an agent
// reading either gets the same 5-way positioning a human reads on the page. No cell contains a
// `|`, so the table needs no escaping.
export function compareMarkdown(): string {
  const head = `| Dimension | ${columns.join(' | ')} |`;
  const sep = `| --- | ${columns.map(() => '---').join(' | ')} |`;
  const body = rows.map((r) => `| ${r.dim} | ${r.cells.join(' | ')} |`).join('\n');
  const faq = faqs.map((f) => `### ${f.q}\n\n${f.a}`).join('\n\n');
  return (
    `A prompt, a skill, an MCP server, a fine-tune — each moves capability between agents. Only a ` +
    `recipe is a self-contained artifact that rebuilds it in a repository that has never seen ` +
    `yours. Where each one fits:\n\n` +
    `${head}\n${sep}\n${body}\n\n` +
    `A recipe does not replace a skill: \`sporo init\` installs the skill that *writes* recipes — ` +
    `the skill authors in your repo, the recipe travels to the next one.\n\n` +
    `## Frequently asked\n\n${faq}`
  );
}

export const faqs = [
  {
    q: 'What is the difference between a recipe and a skill?',
    a: 'A skill runs inside your harness — it is installed into your agent and acts where it lives. A recipe is a file that rebuilds the capability in a harness that has never seen yours. sporo init installs the sporo-recipe skill, which is the thing that authors recipes.',
  },
  {
    q: 'Can’t I just share a prompt or a doc?',
    a: 'A prompt assumes your context and drifts the moment it leaves it. A recipe is a machine-gated genre: a neutrality rule (roles, never your paths or product names), a literal acceptance check on every build step, and scars recorded as symptom → root cause → fix — so an agent in another repository can actually act on it.',
  },
  {
    q: 'What is the difference between a recipe and fine-tuning a model?',
    a: 'Fine-tuning bakes behavior into a model’s weights: you assemble a dataset, train, and host the result, and the behavior lives in that one model. A recipe is a plain markdown file any model can read and act on — the build steps and the scars are written down instead of averaged into weights, so nothing is retrained and no model is locked in.',
  },
  {
    q: 'Do I need an account, a server, or a runtime?',
    a: 'No. sporo is one static binary and a recipe is one markdown file. Zero accounts, zero servers, zero runtimes — install with curl -fsSL sporo.dev/install.sh | sh.',
  },
  {
    q: 'What stacks does a recipe work on?',
    a: 'Any. A recipe names roles — “the facts file”, “the collector” — never your files, commands, or product, so the reader’s agent maps those roles onto its own stack and harness.',
  },
  {
    q: 'How does a recipe improve?',
    a: 'Through report-backs. A reader hits a new scar, records it, and it becomes the recipe’s next sealed version. The report-back loop, not the file format, is what compounds.',
  },
  {
    q: 'When should I not use a recipe?',
    a: 'When the capability is trivial, genuinely one-off, or so tied to your private infrastructure that nothing about it transfers. Then a prompt or an in-house skill is enough — recipes earn their gating when the capability is worth rebuilding elsewhere.',
  },
  {
    q: 'Is sporo open source?',
    a: 'Yes, Apache-2.0. One static binary for macOS, Linux and Windows, with checksummed release archives for six platforms.',
  },
];
