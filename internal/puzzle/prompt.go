package puzzle

// const Prompt = `
// Generate 10 word-association puzzles as a JSON array.

// Each puzzle is an object with exactly two fields:
// {"answer": "...", "clues": ["...", "...", "...", "...", "..."]}

// Rules:
// - Answers should be everyday concepts, objects, phrases or well-known names — not obscure trivia.
// - Answers should have several distinct meanings, or that appear in
//   unrelated contexts. A good answer can be clued from at least three
//   different domains.
// - Do not describe, define, or paraphrase the answer.
// - Each puzzle has exactly 5 unique clues with 1 or 2 words.
// - Each clue is a thing associated with the answer, not a description or synonym of it.
// - Every clue must genuinely relate to the answer.
// - Each clue must be a concrete, specific thing — a noun someone could
//   picture. Never name a category, domain, or field.
// - Order clues from most ambiguous to most specific. Clue 1 should be cryptic and fit many possible
//   answers; clue 5 should be strongly associated with this answer in particular.
// - No clue may contain the answer, any word from the answer, or a plural or variant of it.

// Examples of good puzzles:
// {"answer": "bank", "clues": ["river", "snow", "memory", "blood", "money"]}
// {"answer": "spring", "clues": ["mattress", "season", "water", "coil", "chicken"]}

// Bad (these are definitions, not associations):
// {"answer": "Clock", "clues": ["Time keeper", "Rhythmic pulse", "Face hands", "Ticking sound", "Hour tracker"]},

// Do not use any of these answers: %s

// Respond with only the JSON array. No explanation, no markdown code fences.
// `

const Prompt = `
Generate 5 word-association puzzles.

Return ONLY a valid JSON array. Do not include markdown or any explanation.

Each element must be:

{
  "answer": "...",
  "clues": ["...", "...", "...", "...", "..."]
}

Requirements:

- Each object has exactly two fields: "answer" and "clues".
- "clues" contains exactly 5 unique strings.
- Each clue is 1 or 2 words.
- Answers should be common words, phrases or well-known names.
- Prefer answers that appear in several unrelated contexts or have multiple meanings.

Most important rule:

Clues must be ASSOCIATIONS, never definitions.

Imagine someone saying "This makes me think of the answer."

Good:
bank -> river, snow, memory, blood, money
spring -> mattress, season, water, coil, chicken

Bad:
clock -> time, tick, hands, numbers, wall

Those describe the answer instead of associating with it.

Additional rules:

- Do not use synonyms.
- Do not use the answer, part of the answer, or singular/plural variants of it.
- Every clue must genuinely relate to the answer.
- Use clues from different contexts whenever possible.
- Order clues from least specific to most specific.
- The fifth clue should strongly point to the answer when combined with the previous clues.

Before producing each puzzle, mentally check:

- Are these associations rather than descriptions?
- Are all clues unique?
- Does clue 5 make the answer much easier than clue 1?
- Would a human enjoy solving this puzzle?

Do not use any of these answers: %s

Output only the JSON array.
`
