# Changelog

All notable changes to this project will be documented in this file.

## [0.51.0]

### Added
- Add trust center access requests

## [0.50.1]

### Fixed
- Fix authentification token error

## [0.50.0]

### Added
- Add compliance registries
- Add vendor services

### Chore
- Replace mailhog

## [0.49.0]

### Added
- Add nonconformity registries

### Fixed
- Fix trust center dark mode

## [0.48.1]

### Fixed

- Fix display of download buttons in the public trust center

## [0.48.0]

### Added
- Add baa to vendors
- Add dpa to vendors
- Add contacts to vendors
- Add name to audits
- Add audits to controls

## [0.47.0]

### Added
- Add document draft deletion
- People now have contract start and end dates in the UI and API.
- Lists can filter out people whose contracts have ended.

### Fixed
- Fix closing of document deletion pop up
- Fix creation of empty draft without save

## [0.46.2] - 2025-08-10

### Fixed

- Fix various SQL queries failures due to trust center
- Fix internal information leaking to API


## [0.46.1] - 2025-08-10

### Fixed

- 5xx on risk show page

## [0.46.0] - 2025-08-08

### Added

- Add organization deletion
- Add Probo by default

## [0.45.1] - 2025-08-06

### Fixed

- Fix data page display

## [0.45.0] - 2025-08-06

### Added
- Add trust center
- Add edition of document fields

## [0.44.0] - 2025-07-23

### Fixed
- Fix PDF tables
- Fix display issue on control and framework
- Fix control creation

## [0.43.1] - 2025-07-22

### Fixed
- Fix document draft creation

## [0.43.0] - 2025-07-21

### Added
- Add control exclusion

### Fixed
- Fix small issues on SOA

## [0.42.1] - 2025-07-16

### Fixed
- Fix missing document download button

## [0.42.0] - 2025-07-16

### Fixed
- Fix document version selector
- Fix duplicate people

## [0.41.0] - 2025-07-16

### Changed

- Revision of multiple UI elements

### Added

- Add document version selector on details page
- Add document bulk publication
- Add document bulk signature request
- Add cancel signature request

## [0.40.0] - 2025-07-11

### Added

- Add cancel request mutation
- Add bulk publish document version mutation
- Add bulk request signature mutation

## [0.39.0] - 2025-06-03

### Added

- Add policy PDF export

### Security

- Update go dependencies
- Update node dependencies

## [0.38.1] - 2025-07-03

### Fixed

- Fix 5xx on document type order

## [0.38.0] - 2025-07-03

### Added

- Allow to change doucment order in the UI

### Change

- Change default document sorting order

## [0.37.5] - 2025-06-30

### Fixed

- Fix missing risk score on detail risk page
- Fix matrix risk score color on risk matrix

## [0.37.4] - 2025-06-30

### Fixed

- Fix SOA with risk

## [0.37.3] - 2025-06-30

### Fixed

- Fix missing framework controls

## [0.37.2] - 2025-06-30

### Changed

- Generate excel in memory instead of using fs

## [0.37.1] - 2025-06-30

### Added

- Add updated at and created at order for vendor

### Fixed

- Fix SOA filename

## [0.37.0] - 2025-06-30

### Added

- Add SOA generator
- Show last assessment date

## [0.36.0] - 2025-06-29

### Added

- Add URI evidence type
- Add link dialog for measure evidences
- Add default security header to API
- Add support for extra header

### Fixed

- Fix tasks deadline
- Fix order people by kind
- Fix missing people role order

### Security

- Remove all data after logout
- Enforce maximum password limit
- Mitigate timing attack on signin

### Changed

- Use httplogger on GraphQL error
- Returns internal error when error is known

## [0.35.0] - 2025-06-20

### Added

- Add forgot password pages

## [0.34.0] - 2025-06-17

### Added

- Pagination for people, vendors, documents, data and assets

### Fixed

- Fix 404 on email confirmation page
- Fix 404 on invitation confirmation page
- Fix login redirection
- Fix form not reset after submit

## [0.33.6] - 2025-06-15

### Fixed

- Fix filedrop upload too small file size

## [0.33.5] - 2025-06-14

### Fixed

- Fix framework view too many queries
- Fix image upload failed

## [0.33.4] - 2025-06-13

- Fix measure count

## [0.33.3] - 2025-06-13

### Fixed

- Fix 5xx on document count for risk
- Fix leaking pg connections

## [0.33.2] - 2025-06-13

### Fixed

- Fix API path contain undefined

## [0.33.1] - 2025-06-13

### Fixed

- Fix localhost enforce at build time

## [0.33.0] - 2025-06-13

### Added

- Add backend fulltext search on controls, documents, risks and measures
- Add `totalCount` field in `Connection` object
- New console design

### Fixed

- Use new enum for data classification

## [0.32.0] - 2025-06-07

### Changed

- Prevent publishing of document versions with no changes
- Update AI prompt used for changelog generation

### Added

- Add deadline on tasks
- Add controls manual create, update, and delete

## [0.31.0] - 2025-06-04

### Added

- Add assets inventory
- Add data inventory
- Add title and owner id to document versions
- Add automatic changelog

## [0.30.1] - 2025-06-02

### Added

- Added sort key `updated_at` for vendors

### Fixed

- Fix 5xx on signature request
- Fix sort key 5xx

## [0.30.0] - 2025-05-31

### Changed

- Change policies to documents

### Added

- Add type to documents

### Fixed

- Fix HIPAA import

## [0.29.0] - 2025-05-29

### Added

- HIPAA releated risks
- Add framework import from json

### Changed

- New add vendor UI

### Fixed

- Add url input type for vendor assessement

## [0.28.0] - 2025-05-29

### Added

- Add automatic vendor assessment
- Add vendor category fields
- Add vendor business associate agreement url and subprocessors list url

### Fixed

- Fix 5xx when create new vendor
- Fix 5xx when update a vendor

## [0.27.1] - 2025-05-25

### Fixed

- Fix missing `position` field migration

## [0.27.0] - 2025-05-25

### Changed

- Rename severity into risk score
- Add contract dates fields
- Add people position field

### Fixed

- Fix conflict http header fields

### [0.26.0] - 2025-05-20

### Added

- Add docker image security scan
- Show latest risk updated date
- Log error when GraphQL resolver failed

### Security

- Update Golang dependencies
- Update to latest Ubuntu LTS
- Update to latest Golang version

### [0.25.0] - 2025-05-13

### Changed

- Allow data and text file for evidences

### Fixed

- Fix missing people when inviting user already in other organization
- Fix cannot upload organization logo

### [0.24.0] - 2025-05-06

### Added

- Task page list

### Changed

- Task is now linked to organization

### Fixed

- Fix cannot see vendor assessment note

### Security

- Add filetype validation for end-user upload

## [0.23.1] - 2025-05-04

### Fixed

- Fix http cache etag
- Fix cannot delete measure

## [0.23.0] - 2025-05-04

### Added

- Add ISO 27001 document header
- Add policy downlaod

### Changed

- Enable HTML support in Markdown renderer

## [0.22.0] - 2025-05-04

### Added

- Show owner of the policy in list
- Show number of singatures in the policy list

### Changed

- Link evidences to measure

## [0.21.0] - 2025-05-01

### Changed

- Add markdown table support
- Explicit risk score calcul

## [0.20.1] - 2025-05-01

### Changed

- New vendors in the built-in lists

## [0.20.0] - 2025-05-01

### Added

- Add end-user confirmation before sending policy sign notification
- Add assessed at in the vendor list

### Fixed

- Fix not aligned button on policy list view

### Removed

- Remove start and end service date of vendor

## [0.19.2] - 2025-05-01

### Fixed

- Fix 5xx when invite user in an organization

## [0.19.1] - 2025-05-01

### Fixed

- Evidence URL not set

## [0.19.0] - 2025-04-30

### Changed

- New vendors in the built-in lists

### Security

- Update javascript dependencies
- Fix open redirect when the redirect url use `//`

## [0.18.1] - 2025-04-30

### Fixed

Fix typo `mesure` instead of `measure`

## [0.18.0] - 2025-04-30

### Added

- Static files are served using GZip
- Static fiels are served with ETag and Cache header fields

### Fixed

- Entrypoint JS/CSS has no chunk hash

## [0.17.0] = 2025-04-29

### Added

- Add policy unlogged sign

## [0.16.0] = 2025-04-29

### Added

- Policy history
- Policy signature
- New vendors in the built-in lists

### Fixed

- Fix cannot delete measure with linked risk

## [0.15.1] - 2025-04-27

### Fixed

- Fix SQL syntax error

## [0.15.0] - 2025-04-27

### Added

- Add delete measure in the UI and GraphQL API

### Removed

- Remove `importance` field from measure as it's not used anymore

### Fixed

- Fix delete evidence from task list does not work
- Fix cannot load attached measure risks

## [0.14.0] - 2025-04-24

### Added

- Risk can have note
- Cache static assets

## [0.13.2] - 2025-04-24

### Fixed

- Fix psql `generated_gid` returns padded base64
- Fix `user_id` not set when create new organization
- Fix `additional_email_addresses` not set when invite in organization

## [0.13.0] - 2025-04-23

### Added

- New "Risk assessments" tab for vendors that allows you to:
  - View all risk assessments for a vendor in one place
  - Create new risk assessments with data sensitivity and business impact ratings
  - Track assessment expiration dates
- Automatic people record creation when accepting invitations
- New vendors in the built-in lists
- Introduced a connector framework enabling integration with external
  services:
  - Add OAuth2 connector implementation

### Changed

- Completely redesigned vendor list page
- Completely redesigned vendor detail page
- Improved compliance reports table with better file size formatting and date display
- People may be linked to user

## [0.12.0] - 2025-04-20

### Added

- New vendors in the built-in lists

### Changed

- Update risk library with new risks

### Security

- Upgrade Golang dependencies
- Upgrade Node dependencies

## [0.11.1] - 2025-04-15

### Changed

- New vendors in the built-in vendors list

## [0.11.0] - 2025-04-15

### Added

- New vendors in the built-in vendors list

### Changed

- More explicit scale, legend and score for risk matrix

## [0.10.1] - 2025-04-14

### Fixed

- Fix grammar for "people" in the navigation bar
- Fix editor change cursor position at each keystroke
- Fix editor does not display list icon

## [0.10.0] - 2025-04-14

### Changed

- Improve UI of the risk matrix
- Rename "Mitigation" in "Mesure"

## [0.9.0] - 2025-04-12

### Added

- Added business owner and security owner fields to vendors

### Changed

- Improved vendor detail page with organized sections
  - Split information into logical sections (Basic Information, Ownership, Risk & Service, Documentation)
  - Better visual organization of vendor information

## [0.8.0] - 2025-04-12

### Added

- New risk treatment strategy options: Mitigate, Accept, Avoid, Transfer
- Risk ownership functionality

## [0.7.0] - 2025-04-12

### Added

- Enhanced risk management with inherent and residual risk assessment capabilities
  - Added new fields to track both inherent and residual likelihood/impact values
  - Introduced risk severity calculation as the product of likelihood and impact
  - Added visual risk matrix to view risk distribution by severity
- New risk-policy mapping functionality allowing risks to be linked to policies
- New risk-control mapping functionality enabling risks to be linked to controls
- Added edit functionality for risks with a new edit page
- New popover components for mitigation information on the mitigations list view
- Pre-populated risk templates from a JSON data source

### Changed

- Update vendors catalog.
- Updated risk creation form to include both inherent and residual risk parameters
- Improved risk list view with risk matrix visualization
- Enhanced breadcrumb navigation for risk detail pages
- Refactored risk-mitigation mapping to remove redundant probability/impact fields
- Renamed probability field to likelihood for better alignment with risk management terminology

### Fixed

- Improved license file formatting in vendors and risks data directories
- Fixed URL in attribution text (`getprobo.com` → `www.getprobo.com`)

## [0.6.0] - 2025-04-10

### Added

- Added vendors.json data file under Creative Commons Attribution-ShareAlike 4.0 license`
- New vendor data management system with comprehensive vendor information
- Pre-populated vendor database with 12 common SaaS vendors and their certifications
- Vendor details page with extended fields for improved vendor management:
  - Legal name and headquarters address
  - Website URL
  - Certification tracking with tag-based interface
  - Links to important vendor documents (SLA, DPA, security pages)
  - Support for multiple compliance certifications per vendor

### Fixed

- Fix cannot create vendor when the name is too similar to suggested one
- Fix UI showing double button to close evidence preview modal
- Fix cannot delete vendor with compliance reports (added cascade delete constraint)

## [0.5.0] - 2025-04-10

### Added

- Add vendor compliance reports UI
- Controls can now be linked to policies, enabling better organization of compliance documentation and clearer traceability between policies and security controls
- New UI for viewing and managing policies related to a specific control

## [0.4.2] - 2025-04-09

### Changed

- Simplified policy data model by removing version field and optimistic concurrency
- Refactored policy update flow to load-modify-save pattern

### Fixed

- Added user-friendly error messages when importing frameworks that already exist

## [0.4.1] - 2025-04-09

### Changed

- Update ISO 27001 and SOC2 framework definition.

## [0.4.0] - 2025-04-09

### BREAKING CHANGES

- **BREAKING:** Renamed GraphQL mutations for control-mitigation mappings:
  - `createControlMapping` → `createControlMitigationMapping`
  - `deleteControlMapping` → `deleteControlMitigationMapping`
  - Input and payload types have been updated accordingly

### Added

- Add import control <> mitigation mapping.
- Add mitigation tasks import.
- Add auto-scroll to opened category.
- Added support for mapping controls to policies:
  - New GraphQL mutations `createControlPolicyMapping` and `deleteControlPolicyMapping`
  - Controls can now be associated with both mitigations and policies
  - New bidirectional relationships:
    - Control objects now expose a `policies` field to list associated policies
    - Policy objects now expose a `controls` field to list associated controls
- Added vendor compliance reports:
  - New GraphQL types `VendorComplianceReport` and related connection types
  - New GraphQL mutations `uploadVendorComplianceReport` and `deleteVendorComplianceReport`
  - New `complianceReports` field on the Vendor type
  - Support for uploading, viewing, and managing vendor compliance documentation
- Added pre-configured frameworks:
  - Added ISO/IEC 27001:2022 and SOC 2 framework templates
  - Improved framework import interface with dropdown menu for template selection
  - Support for one-click import of standard compliance frameworks

### Changed

- Evidence can now be requested.

### Fixed

- Fix unfoldable mitigation category when open via the URI fragment.
- Fix ctrl+click on mitigation does not open new tab.
- Fix error handling in framework view when no controls are available.

## [0.3.0] - 2025-04-01

### Added

- Add sidebar to show a task.
- Add task estimate edition.
- Add control+framework auditor views.
- Add import mitigations support.
- Add import framework support.
- Add risk object management.
- Add risk template.
- Add mapping between control and risk.

### Changed

- Rename control in mitigation.
- Home page is now mitigations page.

### Fixed

- Fix panic in GraphQL resolver are not reported.
- Fix otal trace never started.
- Fix React.lazy chunck error.
- Fix login page show `unauthorized` error.
- Fix cannot delete task with evidences.
- Fix cannot download file with non-ASCII filename.

## [0.2.0] - 2025-03-24

### Added

- Add forget password.
- Allow evidence to be a link.
- Add task import support.
- Allow to create vendor when it not exist in the auto-complete.
- Add service account people kind.

### Changed

- Make task time estimate optional.
- Set invitation token to 12 hours.
- Order people by fullname.
- Order vendor by name.
- Allow to edit control state without going to edit page.
- Redirect on people list after people creation.
- New UI for the framework overview page.

### Fixed

- Fix flickering on hover on categories.
- Fix control order under a category.
- Fix UI does not refresh after importing a framework.
- Fix cannot create control.
- Fix missing include cookie on confirmation invit.
- Fix sign-in does not include cookie.
- Fix missing version when create task.
- Fix random order on framework overview.
- Fix change task state not visible on UI.
- Fix control card items alignement.
- Fix cannot delete task.
- Fix password managers misidentifying token fields as usernames in reset password and invitation confirmation forms.

## [0.1.0] - 2025-03-14

Initial release.
