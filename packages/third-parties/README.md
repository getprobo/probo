# ThirdParties library

A curated collection of software thirdParty information for use in thirdParty management and compliance activities.

## Overview

This package provides a structured dataset of software thirdParties and service providers with comprehensive metadata including:

- Basic thirdParty information (name, website, description)
- Legal documentation URLs (privacy policy, terms of service, etc.)
- Compliance certifications
- Security information

## Data Structure

Each thirdParty entry follows the structure defined in `data.d.ts`, including fields for:

- `name`: Display name of the thirdParty
- `legalName`: Legal business name
- `websiteUrl`: ThirdParty's website
- `privacyPolicyUrl`: URL to privacy policy
- `termsOfServiceUrl`: URL to terms of service
- And many more compliance-related URLs and metadata

See `data.d.ts` for the complete type definition.

### Building the Markdown Documentation

To generate the VENDORS.md documentation file from the data:

```bash
npm run build:md
```

## License

CC BY-SA 4.0 - See LICENSE.md for details.
