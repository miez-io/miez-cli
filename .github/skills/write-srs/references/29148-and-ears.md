# 29148 and EARS

Load this when writing or reviewing a requirement.

## ISO/IEC/IEEE 29148

[ISO/IEC/IEEE 29148](https://www.iso.org/standard/72089.html) is the
requirements-engineering standard. It defines what a good requirement is
and what a requirements specification contains. It does not define a
sentence template.

Publicly cited characteristics (29148 / SEVOCAB):

| Characteristic | Meaning |
|---|---|
| Necessary | If removed, a real need is lost |
| Implementation-free | States what, not how |
| Unambiguous | One reading for every reader |
| Complete | All needed conditions are stated |
| Singular | One need |
| Feasible | Can be met with known constraints |
| Verifiable | A test can pass or fail |
| Consistent | Does not fight other requirements |
| Traceable | Stable identifier |

The document that holds these lines is a **software requirements
specification (SRS)**. IEEE 830-1998 defined the common SRS outline and
was superseded by 29148. The outline is still the usual section list:

1. Introduction (purpose, scope, definitions)
2. Overall description (context, users, constraints)
3. Specific requirements
4. Verification
5. Appendices

Quality categories (performance, security, reliability, usability) come
from [ISO/IEC 25010](https://www.iso.org/standard/35733.html). They are
headings inside the SRS, not a second document.

## EARS

[Easy Approach to Requirements Syntax](https://ieeexplore.ieee.org/document/5261575)
(Mavin, Wilkinson, Harwood, Harrison; INCOSE 2009) is the sentence form.

The five patterns:

| Pattern | When to use | Form |
|---|---|---|
| Ubiquitous | Always true | The `<system>` shall `<response>`. |
| Event-driven | Triggered by an event | When `<trigger>`, the `<system>` shall `<response>`. |
| State-driven | While a state holds | While `<state>`, the `<system>` shall `<response>`. |
| Optional | Feature is present | Where `<feature>`, the `<system>` shall `<response>`. |
| Unwanted | Fault or abuse | If `<unwanted>`, the `<system>` shall `<response>`. |

Complex EARS may combine optional, event, and state in one sentence.
Prefer one pattern unless the combination is the actual need.

EARS does not replace 29148. A well-formed EARS line can still pack two
needs or hide design. Apply both.

## Acceptance

29148 requires verifiability. Under each shall-line write a check a test
can pass or fail. A WHEN / THEN line is enough. Words such as fast,
easy, or secure are not a check.

## IDs

Use a stable, speakable ID (`capability/short-name` or the project's
existing scheme). Do not rename an ID when wording changes. Do not reuse
an ID for a different need.

## Sources

- ISO/IEC/IEEE 29148, Systems and software engineering — Life cycle
  processes — Requirements engineering
- ISO/IEC/IEEE 24765 (SEVOCAB) for the characteristic definitions
- Alistair Mavin et al., “Easy Approach to Requirements Syntax (EARS)”,
  INCOSE International Symposium, 2009
- IEEE 830-1998 (historical SRS outline)
