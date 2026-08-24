---
paths:
  - "**/*.go"
  - "**/*.md"
---

# Technical writing style (go-openapi)

Applies to every committed comment, commit message, README and doc-site page.

The standard is Ernest Gowers, *Plain Words*: **be short, be simple, be human.**
His worked example is the whole rule:

    DON'T   Was this the realisation of an anticipated liability?
    DO      Did you expect to have to do this?

The abstract nouns carry no information; the concrete verb carries all of it.

## Two tests

**The grep test.** Does the sentence contain something a reader can search for — an
identifier, a file, a flag, an error, a number with a unit? Prose that names nothing has
described the code without pointing at it.

**The quotability test.** A sentence that would survive being quoted on its own is too
pleased with itself. Rewrite it until it merely sounds true.

## Never define by inversion

The worst and most frequent fault. A copula whose subject or predicate is a wh-clause
promises a definition and delivers a metaphor. Both directions are banned:

    DON'T   Coverage is what says which templates a suite never reaches.
    DON'T   What is lost is the doc comment.
    DO      Coverage records which templates the suite never executed.
    DO      A synthesized type loses its doc comment.

The rewrite is mechanical: find the verb hiding inside the wh-clause and make it the main
verb of the sentence.

`which is why` pointing back at a fact just stated is legitimate, and rationed — one per
comment is plenty.

## The rest

- **Name the thing.** `WithRoots`, not "the option that scopes a repository". Name the
  error, the file, the flag, the upstream package, the constant.
- **Statement, not aphorism.** State mechanism and effect. Never close a paragraph on a
  maxim: the reflex lands hardest on a closing sentence.
- **Keep a subject.** "New returns an error if the source is unreadable", not "What a
  source leaves out is settled where it is declared".
- **Plain verbs.** add, fix, return, parse, reject, cap, prune, record. Code does not say,
  judge, grant, refuse, know, mean to, or reach for. `report` is fine when something
  genuinely reports.
- **Keep the numbers.** Sizes with units, counts, ratios, advisory ids. `286 -> 178 KiB`,
  `GHSA-v2xp-g8xf-22pf`. Dropping them for a smoother sentence loses information.
- **Be human.** Address the reader where there is advice: "Use `WithRoot` to confine local
  loading." Admit the awkward thing rather than smoothing it over.

## Self-check

    # definition by inversion, both directions
    grep -rnE '\b(is|are) (what|where) [a-z]' --include='*.go' --include='*.md' .
    grep -rnE '(^|\. )What [a-z][a-z ,-]{3,50} (is|are) ' --include='*.go' --include='*.md' .

Subtract the legitimate `which/that/this/it is what` before judging the first one.
Neither grep is a verdict — they find one fault out of six. The others need reading.
