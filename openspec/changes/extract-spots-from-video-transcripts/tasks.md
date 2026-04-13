## 1. Video detection and transcript acquisition

- [ ] 1.1 Extend article ingestion to detect embedded YouTube videos needed for transcript-first extraction
- [ ] 1.2 Add transcript acquisition and parsing with timestamp preservation

## 2. Transcript-first extraction

- [ ] 2.1 Implement transcript-based structured extraction for author, spots, addresses, and timestamps
- [ ] 2.2 Ensure the current pipeline no longer depends on the removed article-text extraction path

## 3. Persistence changes

- [ ] 3.1 Extend persistence/storage to support source video metadata for transcript-derived spots
- [ ] 3.2 Keep exported/frontend-consumable spot data coherent with the new storage model

## 4. Verification

- [ ] 4.1 Verify embedded-video articles can reach transcript extraction successfully
- [ ] 4.2 Verify extracted timestamps are suitable for later YouTube deep-linking
