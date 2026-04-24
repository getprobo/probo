# Vendors library

A curated collection of software vendor information for use in vendor management and compliance activities.

## Overview

This package provides a structured dataset of software vendors and service providers with comprehensive metadata including:

- Basic vendor information (name, website, description)
- Legal documentation URLs (privacy policy, terms of service, etc.)
- Compliance certifications
- Security information

## Data Structure

Each vendor entry follows the structure defined in `data.d.ts`, including fields for:

- `name`: Display name of the vendor
- `legalName`: Legal business name
- `websiteUrl`: Vendor's website
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
