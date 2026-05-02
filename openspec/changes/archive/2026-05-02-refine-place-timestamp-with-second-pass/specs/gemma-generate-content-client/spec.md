## MODIFIED Requirements

### Requirement: Request payload carries Dutch prompt text as model input
The client SHALL send Dutch prompt content in the request body using the API's content/parts text structure for both extraction passes.

#### Scenario: Pass-1 prompt included in request
- **WHEN** the extraction command invokes the model client for pass 1
- **THEN** the request body SHALL include the composed Dutch pass-1 prompt as text content
- **AND** the pass-1 prompt SHALL include cleaned article text and transcript sentence units as separate labeled sections
- **AND** the pass-1 prompt SHALL explicitly state that article text is primary for place identification and transcript units are required for mention timing/evidence
- **AND** the request SHALL target the configured Gemma model identifier

#### Scenario: Pass-2 batched refinement prompt included in request
- **WHEN** the extraction command invokes the model client for pass 2 refinement
- **THEN** the request body SHALL include a Dutch pass-2 prompt as text content
- **AND** the pass-2 prompt SHALL include the full set of pass-1 place/timestamp results plus transcript sentence units
- **AND** the pass-2 prompt SHALL request refined timestamps for all places in one batched response
- **AND** the pass-2 prompt SHALL enforce that refinement is earlier-or-equal to pass-1 timestamps and aligned to transcript evidence
- **AND** if any pass-2 refinement cannot be validated against transcript evidence or violates timestamp constraints, the system SHALL retain the original pass-1 timestamp for that place
- **AND** the request SHALL target the configured Gemma model identifier
