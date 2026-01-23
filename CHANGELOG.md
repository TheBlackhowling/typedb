# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]


## [0.1.31] - 2026-01-23

**Commit:** [`055c664`](https://github.com/TheBlackhowling/typedb/commit/055c664e21def76f5ea90b697ce2d8a3e3cc78e7)
**Pull Request:** [#38](https://github.com/TheBlackhowling/typedb/pull/38)

**Summary:** Add validation and quote escaping to quoteIdentifier function to prevent identifier injection

**Key Changes:**
- See detailed changes below

**Detailed Changes:** See [versions/0.1.31.md](versions/0.1.31.md)


## [0.1.30] - 2026-01-23

**Commit:** [`460cdc3`](https://github.com/TheBlackhowling/typedb/commit/460cdc375b4b9b1259800a19d6c3eb9ecb525b62)
**Pull Request:** [#37](https://github.com/TheBlackhowling/typedb/pull/37)

**Summary:** Fix SQL injection vulnerability in Oracle InsertAndReturn by validating and escaping identifiers

**Key Changes:**
- See detailed changes below

**Detailed Changes:** See [versions/0.1.30.md](versions/0.1.30.md)


## [0.1.29] - 2026-01-20

**Commit:** [`d1a287d`](https://github.com/TheBlackhowling/typedb/commit/d1a287de5533c1a1b9fb99c16721514f1db7ee24)
**Pull Request:** [#36](https://github.com/TheBlackhowling/typedb/pull/36)

**Summary:** Add validation during RegisterModel and RegisterModelWithOptions to catch missing QueryBy methods early and fail fast at registration time

**Key Changes:**
- See detailed changes below

**Detailed Changes:** See [versions/0.1.29.md](versions/0.1.29.md)


## [0.1.28] - 2026-01-20

**Commit:** [`fc769f3`](https://github.com/TheBlackhowling/typedb/commit/fc769f3ef5bf04b64ffc5fc7cbd072fd2cea1fae)
**Pull Request:** [#35](https://github.com/TheBlackhowling/typedb/pull/35)

**Summary:** Improve error message when LoadByComposite is called but required QueryBy method is missing to provide better developer experience

**Key Changes:**
- See detailed changes below

**Detailed Changes:** See [versions/0.1.28.md](versions/0.1.28.md)


## [0.1.27] - 2026-01-20

**Commit:** [`c03da1d`](https://github.com/TheBlackhowling/typedb/commit/c03da1d630f126ab94eb8c3a1f2de8a74f787edf)
**Pull Request:** [#34](https://github.com/TheBlackhowling/typedb/pull/34)

**Summary:** Extract shared field iteration logic from serializeModelFields, serializeModelFieldsForUpdate, and buildFieldMapForComparison into a reusable helper function

**Key Changes:**
- See detailed changes below

**Detailed Changes:** See [versions/0.1.27.md](versions/0.1.27.md)


## [0.1.26] - 2026-01-20

**Commit:** [`4c45cf3`](https://github.com/TheBlackhowling/typedb/commit/4c45cf34187c575b96d5de7655d538dbf63f694e)
**Pull Request:** [#33](https://github.com/TheBlackhowling/typedb/pull/33)

**Summary:** Refactor Open() and OpenWithoutValidation() to eliminate code duplication by extracting shared connection setup logic

**Key Changes:**
- See detailed changes below

**Detailed Changes:** See [versions/0.1.26.md](versions/0.1.26.md)


## [0.1.25] - 2026-01-20

**Commit:** [`2dd2464`](https://github.com/TheBlackhowling/typedb/commit/2dd2464772f5c6176eb66aee1d5d3176bb4b1b7b)
**Pull Request:** [#32](https://github.com/TheBlackhowling/typedb/pull/32)

**Summary:** Refactor executor.go to eliminate code duplication between DB and Tx methods by extracting shared helper functions

**Key Changes:**
- See detailed changes below

**Detailed Changes:** See [versions/0.1.25.md](versions/0.1.25.md)


## [0.1.24] - 2026-01-20

**Commit:** [`052a9c7`](https://github.com/TheBlackhowling/typedb/commit/052a9c769eb7c370f723042bd4188a802ea93674)
**Pull Request:** [#31](https://github.com/TheBlackhowling/typedb/pull/31)

**Summary:** Document when bulk queries (including 50K+ rows) are reasonable to use typedb, explaining that reflection overhead scales linearly with row count

**Key Changes:**
- See detailed changes below

**Detailed Changes:** See [versions/0.1.24.md](versions/0.1.24.md)


## [0.1.23] - 2026-01-20

**Commit:** [`9e29b69`](https://github.com/TheBlackhowling/typedb/commit/9e29b69c62fc6f26dddd5f9d166d5fde083e0404)
**Pull Request:** [#30](https://github.com/TheBlackhowling/typedb/pull/30)

**Summary:** Add pluggable logging interface to typedb for debugging and monitoring

**Key Changes:**
- See detailed changes below

**Detailed Changes:** See [versions/0.1.23.md](versions/0.1.23.md)


## [0.1.22] - 2026-01-19

**Commit:** [`2a49cca`](https://github.com/TheBlackhowling/typedb/commit/2a49cca2eaa21d2df3fc6b89516381063e313372)
**Pull Request:** [#29](https://github.com/TheBlackhowling/typedb/pull/29)

**Summary:** Enhance README with better structure, clearer value proposition, and improved navigation

**Key Changes:**
- See detailed changes below

**Detailed Changes:** See [versions/0.1.22.md](versions/0.1.22.md)


## [0.1.21] - 2026-01-19

**Commit:** [`382f6c9`](https://github.com/TheBlackhowling/typedb/commit/382f6c9e95d284d25596ca5c0f1709ce680c593b)
**Pull Request:** [#28](https://github.com/TheBlackhowling/typedb/pull/28)

**Summary:** Fix PostgreSQL examples by adding missing updated_at column migration

**Key Changes:**
- See detailed changes below

**Detailed Changes:** See [versions/0.1.21.md](versions/0.1.21.md)


## [0.1.20] - 2026-01-18

**Commit:** [`ff13db2`](https://github.com/TheBlackhowling/typedb/commit/ff13db22a6a507e730be6860a3c011020b8ca8aa)
**Pull Request:** [#27](https://github.com/TheBlackhowling/typedb/pull/27)

**Summary:** Refactor integration tests into smaller files and reorganize README documentation

**Key Changes:**
- ... and 19 more files

**Detailed Changes:** See [versions/0.1.20.md](versions/0.1.20.md)


## [0.1.19] - 2026-01-18

**Commit:** [`a6eee08`](https://github.com/TheBlackhowling/typedb/commit/a6eee0849afb5b8666385ae0916ae378d3c0945b)
**Pull Request:** [#26](https://github.com/TheBlackhowling/typedb/pull/26)

**Summary:** PR #26

**Key Changes:**
- Removed: Model.Load()

**Detailed Changes:** See [versions/0.1.19.md](versions/0.1.19.md)


## [0.1.18] - 2026-01-18

**Commit:** [`63f0ee6`](https://github.com/TheBlackhowling/typedb/commit/63f0ee6c51d353ad48350fd111808795519102c0)
**Pull Request:** [#24](https://github.com/TheBlackhowling/typedb/pull/24)

**Summary:** Fix compilation error in PostgreSQL examples by adding missing UpdatedAt field

**Key Changes:**
- See detailed changes below

**Detailed Changes:** See [versions/0.1.18.md](versions/0.1.18.md)


## [0.1.17] - 2026-01-18

**Commit:** [`155f0a8`](https://github.com/TheBlackhowling/typedb/commit/155f0a861c2a2653f25b4311c1ed04670e3ae0a7)
**Pull Request:** [#23](https://github.com/TheBlackhowling/typedb/pull/23)

**Summary:** Add partial update feature that tracks changes and only updates modified fields

**Key Changes:**
- ... and 5 more files

**Detailed Changes:** See [versions/0.1.17.md](versions/0.1.17.md)


## [0.1.16] - 2026-01-18

**Commit:** [`00d3d3e`](https://github.com/TheBlackhowling/typedb/commit/00d3d3ed85d5f58753f60c8f32f8a548ed212744)
**Pull Request:** [#22](https://github.com/TheBlackhowling/typedb/pull/22)

**Summary:** Add automatic timestamp population for updated_at fields during UPDATE operations using database functions

**Key Changes:**
- See detailed changes below

**Detailed Changes:** See [versions/0.1.16.md](versions/0.1.16.md)


## [0.1.15] - 2026-01-18

**Commit:** [`babb7f4`](https://github.com/TheBlackhowling/typedb/commit/babb7f4ec141e54bc70cf511a71d86edf7bce4d9)
**Pull Request:** [#21](https://github.com/TheBlackhowling/typedb/pull/21)

**Summary:** Complete comprehensive integration test coverage for all supported databases with full API testing and resilient test patterns

**Key Changes:**
- ... and 2 more files
- ✅ **Automated Workflow**: New GitHub Actions workflow added
- ✅ **Source Code**: Code changes included
- ✅ **Documentation**: Comprehensive documentation added

**Detailed Changes:** See [versions/0.1.15.md](versions/0.1.15.md)


## [0.1.14] - 2026-01-12

**Commit:** [`120d8f4`](https://github.com/TheBlackhowling/typedb/commit/120d8f459ee778bb4a5c0d1cda3d1463fce79734)
**Pull Request:** [#20](https://github.com/TheBlackhowling/typedb/pull/20)

**Summary:** Disable example workflows on PR sync and fix SQLite CI migration issues

**Key Changes:**
- ✅ **Automated Workflow**: New GitHub Actions workflow added
- ✅ **Source Code**: Code changes included

**Detailed Changes:** See [versions/0.1.14.md](versions/0.1.14.md)


## [0.1.13] - 2026-01-12

**Commit:** [`7f09938`](https://github.com/TheBlackhowling/typedb/commit/7f0993867689d58620af14ff01f0c8775746d790)
**Pull Request:** [#19](https://github.com/TheBlackhowling/typedb/pull/19)

**Summary:** Add comprehensive examples and CI workflows for all supported database types with refactored structure

**Key Changes:**
- ... and 4 more files
- ✅ **Automated Workflow**: New GitHub Actions workflow added
- ✅ **Source Code**: Code changes included
- ✅ **Documentation**: Comprehensive documentation added

**Detailed Changes:** See [versions/0.1.13.md](versions/0.1.13.md)


## [0.1.12] - 2026-01-12

**Commit:** [`5dc06ec`](https://github.com/TheBlackhowling/typedb/commit/5dc06ec3ca44b2b71ab88ac428191645b90465e1)
**Pull Request:** [#18](https://github.com/TheBlackhowling/typedb/pull/18)

**Summary:** Remove redundant Deserialize method overrides and improve Model.Deserialize() code coverage

**Key Changes:**
- ✅ **Source Code**: Code changes included

**Detailed Changes:** See [versions/0.1.12.md](versions/0.1.12.md)


## [0.1.11] - 2026-01-11

**Commit:** [`c51b08c`](https://github.com/TheBlackhowling/typedb/commit/c51b08c2f669c186068fc534de42ccd189341032)
**Pull Request:** [#17](https://github.com/TheBlackhowling/typedb/pull/17)

**Summary:** Add Insert by Object feature with comprehensive test coverage and documentation

**Key Changes:**
- ✅ **Source Code**: Code changes included
- ✅ **Documentation**: Comprehensive documentation added

**Detailed Changes:** See [versions/0.1.11.md](versions/0.1.11.md)


## [0.1.10] - 2026-01-11

**Commit:** [`5fdf501`](https://github.com/TheBlackhowling/typedb/commit/5fdf5010901f99cbea05717bea1340ebd154e6b8)
**Pull Request:** [#16](https://github.com/TheBlackhowling/typedb/pull/16)

**Summary:** Add insert functionality with ID retrieval and comprehensive unit tests for insert and uint deserialization functions

**Key Changes:**
- ✅ **Source Code**: Code changes included
- ✅ **Documentation**: Comprehensive documentation added

**Detailed Changes:** See [versions/0.1.10.md](versions/0.1.10.md)


## [0.1.9] - 2026-01-11

**Commit:** [`030933c`](https://github.com/TheBlackhowling/typedb/commit/030933c4810e1520084bce44067cc766ef5820e7)
**Pull Request:** [#15](https://github.com/TheBlackhowling/typedb/pull/15)

**Summary:** Implement Load methods for loading models by primary key, unique fields, and composite keys

**Key Changes:**
- ✅ **Source Code**: Code changes included
- ✅ **Documentation**: Comprehensive documentation added

**Detailed Changes:** See [versions/0.1.9.md](versions/0.1.9.md)


## [0.1.8] - 2026-01-11

**Commit:** [`684771d`](https://github.com/TheBlackhowling/typedb/commit/684771d3a04f412a1a612e9270089e925125a302)
**Pull Request:** [#14](https://github.com/TheBlackhowling/typedb/pull/14)

**Summary:** Add comprehensive sqlmock tests for Layer 4 executor methods to improve code coverage

**Key Changes:**
- ✅ **Source Code**: Code changes included
- ✅ **Documentation**: Comprehensive documentation added

**Detailed Changes:** See [versions/0.1.8.md](versions/0.1.8.md)


## [0.1.7] - 2026-01-11

**Commit:** [`f892cc1`](https://github.com/TheBlackhowling/typedb/commit/f892cc1fbc521298978a8f3e0f27e6975f91e49d)
**Pull Request:** [#13](https://github.com/TheBlackhowling/typedb/pull/13)

**Summary:** Implement Layer 6 query helper functions with comprehensive test coverage

**Key Changes:**
- ✅ **Source Code**: Code changes included
- ✅ **Documentation**: Comprehensive documentation added

**Detailed Changes:** See [versions/0.1.7.md](versions/0.1.7.md)


## [0.1.6] - 2026-01-11

**Commit:** [`a7416d3`](https://github.com/TheBlackhowling/typedb/commit/a7416d3a8aed4d639e02b69133095feabfaa4d43)
**Pull Request:** [#12](https://github.com/TheBlackhowling/typedb/pull/12)

**Summary:** Implement model validation system to ensure registered models have required query methods for load tags

**Key Changes:**
- ✅ **Source Code**: Code changes included

**Detailed Changes:** See [versions/0.1.6.md](versions/0.1.6.md)


## [0.1.5] - 2026-01-11

**Commit:** [`12bef49`](https://github.com/TheBlackhowling/typedb/commit/12bef49dc9efd41aa20ad71fbca445ed650d720a)
**Pull Request:** [#11](https://github.com/TheBlackhowling/typedb/pull/11)

**Summary:** Implement Layer 4 executor interface with database connection management and query execution methods

**Key Changes:**
- ✅ **Source Code**: Code changes included
- ✅ **Documentation**: Comprehensive documentation added

**Detailed Changes:** See [versions/0.1.5.md](versions/0.1.5.md)


## [0.1.4] - 2026-01-11

**Commit:** [`201092c`](https://github.com/TheBlackhowling/typedb/commit/201092c23ea569f0092b7a532048d3590547ed54)
**Pull Request:** [#10](https://github.com/TheBlackhowling/typedb/pull/10)

**Summary:** Implement comprehensive deserialization and serialization layer with 96.7% test coverage

**Key Changes:**
- ✅ **Automated Workflow**: New GitHub Actions workflow added
- ✅ **Source Code**: Code changes included

**Detailed Changes:** See [versions/0.1.4.md](versions/0.1.4.md)


## [0.1.3] - 2026-01-11

**Commit:** [`459b1fe`](https://github.com/TheBlackhowling/typedb/commit/459b1fe8186f803070b6ecc07a43ad545a095b99)
**Pull Request:** [#9](https://github.com/TheBlackhowling/typedb/pull/9)

**Summary:** Implement core model registration and reflection utilities with pointer type enforcement

**Key Changes:**
- ✅ **Source Code**: Code changes included
- ✅ **Documentation**: Comprehensive documentation added

**Detailed Changes:** See [versions/0.1.3.md](versions/0.1.3.md)


## [0.1.2] - 2026-01-11

**Commit:** [`42d3084`](https://github.com/TheBlackhowling/typedb/commit/42d3084eed78975a033718efcb0d16ab7fd05613)
**Pull Request:** [#8](https://github.com/TheBlackhowling/typedb/pull/8)

**Summary:** Add automated testing workflow to run tests and generate coverage reports

**Key Changes:**
- ✅ **Automated Workflow**: New GitHub Actions workflow added
- ✅ **Source Code**: Code changes included
- ✅ **Documentation**: Comprehensive documentation added

**Detailed Changes:** See [versions/0.1.2.md](versions/0.1.2.md)


## [0.1.1] - 2026-01-11

**Commit:** [`5dd4c61`](https://github.com/TheBlackhowling/typedb/commit/5dd4c61a8aaa8f04033f95d554ce6749e9c9f6a5)
**Pull Request:** [#7](https://github.com/TheBlackhowling/typedb/pull/7)

**Summary:** Add GitHub App authentication support for bypassing branch protection when pushing changelog updates

**Key Changes:**
- ✅ **Automated Workflow**: New GitHub Actions workflow added

**Detailed Changes:** See [versions/0.1.1.md](versions/0.1.1.md)

### Added
- Initial project setup
