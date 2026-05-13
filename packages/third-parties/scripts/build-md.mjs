import { readFile } from "node:fs/promises";
import { createWriteStream } from "node:fs";
import path from "node:path";

const data = await readFile(path.join(import.meta.dirname, '../data.json'), 'utf8');
const thirdParties = JSON.parse(data);

const output = path.join(import.meta.dirname, '../VENDORS.md');
const file = createWriteStream(output);

const formatAsList = (array) => {
  if (!array || array.length === 0) return '';
  if (typeof array === 'string') {
    return array.split(',').map(item => `- ${item.trim()}`).join('\n');
  }
  return array.map(item => `- ${item}`).join('\n');
};

file.write('# ThirdParties\n\n');
file.write('## Table of Contents by Category\n\n');

const categoriesMap = new Map();

for (const thirdParty of thirdParties) {
  const category = (thirdParty.category || thirdParty.categories || 'Uncategorized');
  
  if (!categoriesMap.has(category)) {
    categoriesMap.set(category, []);
  }
  
  categoriesMap.get(category).push(thirdParty.name);
}

const sortedCategories = [...categoriesMap.keys()].sort();

for (const category of sortedCategories) {
  file.write(`### ${category}\n\n`);
  
  const thirdPartiesInCategory = categoriesMap.get(category).sort();
  for (const thirdPartyName of thirdPartiesInCategory) {
    // Create proper anchor by:
    // 1. Converting to lowercase
    // 2. Replacing spaces with hyphens
    // 3. Removing parentheses, dots, and other special characters
    const anchor = thirdPartyName.toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[\(\)\.]/g, '')
      .replace(/[^a-z0-9\-]/g, '');
    
    file.write(`- [${thirdPartyName}](#${anchor})\n`);
  }
  
  file.write('\n');
}

file.write('---\n\n');

for (const thirdParty of thirdParties) {
  file.write(`## ${thirdParty.name}\n\n`);
  if (thirdParty.description) {
    file.write(`${thirdParty.description}\n\n`);
  }
  
  if (thirdParty.legalName) {
    file.write(`**Legal Name:** ${thirdParty.legalName}\n\n`);
  }
  
  if (thirdParty.headquarterAddress) {
    file.write(`**Headquarters:** ${thirdParty.headquarterAddress}\n\n`);
  }
  
  file.write('### Links\n\n');
  file.write('| Resource | Link |\n');
  file.write('|----------|------|\n');
  
  if (thirdParty.websiteUrl) {
    file.write(`| Website | [Link](${thirdParty.websiteUrl}) |\n`);
  }
  
  if (thirdParty.privacyPolicyUrl) {
    file.write(`| Privacy Policy | [Link](${thirdParty.privacyPolicyUrl}) |\n`);
  }
  
  if (thirdParty.termsOfServiceUrl && thirdParty.termsOfServiceUrl !== 'undefined') {
    file.write(`| Terms of Service | [Link](${thirdParty.termsOfServiceUrl}) |\n`);
  }
  
  if (thirdParty.serviceLevelAgreementUrl && thirdParty.serviceLevelAgreementUrl !== 'undefined') {
    file.write(`| Service Level Agreement | [Link](${thirdParty.serviceLevelAgreementUrl}) |\n`);
  }
  
  if (thirdParty.securityPageUrl && thirdParty.securityPageUrl !== 'undefined') {
    file.write(`| Security Page | [Link](${thirdParty.securityPageUrl}) |\n`);
  }
  
  if (thirdParty.trustPageUrl && thirdParty.trustPageUrl !== 'undefined') {
    file.write(`| Trust Page | [Link](${thirdParty.trustPageUrl}) |\n`);
  }
  
  if (thirdParty.statusPageUrl && thirdParty.statusPageUrl !== 'undefined') {
    file.write(`| Status Page | [Link](${thirdParty.statusPageUrl}) |\n`);
  }
  
  if (thirdParty.dataProcessingAgreementUrl && thirdParty.dataProcessingAgreementUrl !== 'undefined') {
    file.write(`| Data Processing Agreement | [Link](${thirdParty.dataProcessingAgreementUrl}) |\n`);
  }
  
  if (thirdParty.businessAssociateAgreementUrl && thirdParty.businessAssociateAgreementUrl !== 'undefined') {
    file.write(`| Business Associate Agreement | [Link](${thirdParty.businessAssociateAgreementUrl}) |\n`);
  }
  
  if (thirdParty.serviceSoftwareAgreementUrl && thirdParty.serviceSoftwareAgreementUrl !== 'undefined') {
    file.write(`| Service Software Agreement | [Link](${thirdParty.serviceSoftwareAgreementUrl}) |\n`);
  }
  
  if (thirdParty.subprocessorsListUrl && thirdParty.subprocessorsListUrl !== 'undefined') {
    file.write(`| Subprocessors List | [Link](${thirdParty.subprocessorsListUrl}) |\n`);
  }
  
  file.write('\n');
  
  if (thirdParty.categories && thirdParty.categories !== 'undefined') {
    file.write(`**Categories:** ${thirdParty.categories}\n\n`);
  } else if (thirdParty.category && thirdParty.category !== 'undefined') {
    file.write(`**Category:** ${thirdParty.category}\n\n`);
  }
  
  if (thirdParty.certifications && thirdParty.certifications !== 'undefined') {
    file.write('### Certifications\n\n');
    file.write(formatAsList(thirdParty.certifications));
    file.write('\n\n');
  }
  
  if (thirdParty.subprocessors && thirdParty.subprocessors !== 'undefined') {
    file.write('### Subprocessors\n\n');
    file.write(formatAsList(thirdParty.subprocessors));
    file.write('\n\n');
  }
  
  file.write('---\n\n');
}

file.end();
