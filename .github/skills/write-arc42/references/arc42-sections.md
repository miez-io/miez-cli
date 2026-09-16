# arc42 sections

Official section purposes from [docs.arc42.org](https://docs.arc42.org/home/).

| # | Section | Write |
|---|---|---|
| 1 | Introduction and Goals | Requirements overview, stakeholders, top quality goals |
| 2 | Constraints | Technical and organizational limits, conventions |
| 3 | Context and Scope | Business and technical context, external interfaces |
| 4 | Solution Strategy | Fundamental solution decisions and ideas |
| 5 | Building Block View | Static decomposition; black boxes then white boxes |
| 6 | Runtime View | How building blocks interact in important scenarios |
| 7 | Deployment View | Infrastructure and where software runs |
| 8 | Crosscutting Concepts | Recurring approaches (auth, errors, logging) |
| 9 | Architecture Decisions | Important, expensive, risky, or contested choices |
| 10 | Quality Requirements | Quality overview and quality scenarios |
| 11 | Risks and Technical Debt | Known problems and debt |
| 12 | Glossary | Business and technical terms |

Section 5 is the map of containers. Component internals belong in an
SDD, not here. Section 9 is an index plus short pointers. The justified
choice lives in an ADR.

## 42010

[ISO/IEC/IEEE 42010](https://www.iso.org/standard/74393.html) defines an
architecture description as views of a system of interest. arc42 is a
template for that description. It is not a second standard for
requirements or for detailed design.

## C4

[C4 model](https://c4model.com/) (Simon Brown):

| Level | Use in arc42 |
|---|---|
| Context | Section 3 |
| Container | Section 5 |
| Deployment | Section 7 |
| Dynamic | Section 6 |
| Component | SDD, not arc42 |

## Sources

- [arc42 template](https://docs.arc42.org/)
- [arc42 GitHub](https://github.com/arc42/arc42-template)
- ISO/IEC/IEEE 42010, Software, systems and enterprise — Architecture
  description
- [C4 model](https://c4model.com/)
