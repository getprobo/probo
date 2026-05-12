import { IconBook } from "@probo/ui";

import type { DocumentTemplate } from "./types";

export const dataClassificationTemplate: DocumentTemplate = {
  id: "data-classification-policy",
  icon: IconBook,
  name: "Data Classification",
  documentType: "POLICY",
  classification: "INTERNAL",
  title: "Data Classification Policy",
  content: [
    "<h1>📊 Data Classification Policy</h1>",
    "<p></p>",

    "<table><thead><tr><th>Property</th><th>Details</th></tr></thead><tbody>",
    "<tr><td>📋 <strong>Document Owner</strong></td><td>Data Governance Team</td></tr>",
    "<tr><td>🏷️ <strong>Classification</strong></td><td>Internal</td></tr>",
    "<tr><td>📅 <strong>Effective Date</strong></td><td><em>[Enter date]</em></td></tr>",
    "<tr><td>🔄 <strong>Next Review</strong></td><td><em>[Enter date + 12 months]</em></td></tr>",
    "<tr><td>📌 <strong>Version</strong></td><td>1.0</td></tr>",
    "<tr><td>✅ <strong>Approved By</strong></td><td><em>[Name, Title]</em></td></tr>",
    "</tbody></table>",
    "<p></p>",
    "<hr>",
    "<p></p>",

    // 1 · Purpose
    "<h2>1 · Purpose</h2>",
    "<p>This policy defines how <strong>[Organization Name]</strong> classifies, labels, handles, and protects information assets based on their sensitivity and business value. Consistent classification ensures that data receives the appropriate level of protection throughout its lifecycle.</p>",
    "<p></p>",

    // 2 · Classification Levels
    "<h2>2 · Classification Levels</h2>",
    "<p></p>",
    "<table><thead><tr><th>Level</th><th>Label</th><th>Description</th><th>Impact if Disclosed</th></tr></thead><tbody>",
    "<tr><td>🟢</td><td><strong>Public</strong></td><td>Information intended for public consumption</td><td>None — already public</td></tr>",
    "<tr><td>🔵</td><td><strong>Internal</strong></td><td>General business information not meant for the public</td><td>Low — minimal business impact</td></tr>",
    "<tr><td>🟠</td><td><strong>Confidential</strong></td><td>Sensitive business data, PII, financial records</td><td>High — regulatory fines, competitive harm</td></tr>",
    "<tr><td>🔴</td><td><strong>Secret</strong></td><td>Highly sensitive data — trade secrets, key material, PHI</td><td>Critical — severe legal, financial, reputational damage</td></tr>",
    "</tbody></table>",
    "<p></p>",
    "<blockquote><p>📌 <strong>Default Rule:</strong> When in doubt, classify data one level <strong>higher</strong> until a formal review is completed.</p></blockquote>",
    "<p></p>",

    // 3 · Handling Matrix
    "<h2>3 · Handling Matrix</h2>",
    "<p></p>",
    "<table><thead><tr><th>Handling Area</th><th>🟢 Public</th><th>🔵 Internal</th><th>🟠 Confidential</th><th>🔴 Secret</th></tr></thead><tbody>",
    "<tr><td><strong>Storage</strong></td><td>Any system</td><td>Approved systems</td><td>Encrypted at rest (AES-256)</td><td>HSM / dedicated vault</td></tr>",
    "<tr><td><strong>Transmission</strong></td><td>No restrictions</td><td>TLS 1.2+</td><td>TLS 1.2+ · Encrypted email</td><td>End-to-end encrypted channels only</td></tr>",
    "<tr><td><strong>Access Control</strong></td><td>Open</td><td>Authenticated users</td><td>Role-based · Need-to-know</td><td>Named individuals · MFA required</td></tr>",
    "<tr><td><strong>Sharing</strong></td><td>Unrestricted</td><td>Internal only</td><td>Approved recipients · NDA required</td><td>Executive approval per instance</td></tr>",
    "<tr><td><strong>Printing</strong></td><td>Unrestricted</td><td>Standard printers</td><td>Secure print release only</td><td>Prohibited unless approved</td></tr>",
    "<tr><td><strong>Disposal</strong></td><td>Standard delete</td><td>Standard delete</td><td>Secure wipe / shredding</td><td>Certified destruction with audit log</td></tr>",
    "</tbody></table>",
    "<p></p>",

    // 4 · Classification Process
    "<h2>4 · Classification Process</h2>",
    "<p>All new data assets must go through the following 5-step process:</p>",
    "<ol>",
    "<li><strong>🔍 Identify</strong> — Catalog the data asset (name, location, owner, data types contained)</li>",
    "<li><strong>📋 Classify</strong> — Assign a classification level using the matrix in Section 2</li>",
    "<li><strong>🏷️ Label</strong> — Apply classification labels per the labeling guide in Section 5</li>",
    "<li><strong>🔒 Handle</strong> — Apply controls matching the handling matrix in Section 3</li>",
    "<li><strong>🔄 Review</strong> — Re-evaluate classification annually or after significant changes</li>",
    "</ol>",
    "<p></p>",

    // 5 · Labeling Guide
    "<h2>5 · Labeling Guide</h2>",
    "<p></p>",
    "<table><thead><tr><th>Asset Type</th><th>How to Label</th><th>Example</th></tr></thead><tbody>",
    "<tr><td>📄 Documents</td><td>Header + footer on every page</td><td><code>CONFIDENTIAL — [Org Name]</code></td></tr>",
    "<tr><td>📧 Emails</td><td>Subject line prefix</td><td><code>[INTERNAL] Q3 Financials</code></td></tr>",
    "<tr><td>💾 Databases</td><td>Metadata tag on schema/table</td><td><code>classification: secret</code></td></tr>",
    "<tr><td>☁️ Cloud Storage</td><td>Object tag / folder naming</td><td><code>/confidential/hr/records/</code></td></tr>",
    "<tr><td>🗂️ Code Repos</td><td>README badge or .classification file</td><td><code>Classification: Internal</code></td></tr>",
    "</tbody></table>",
    "<p></p>",

    // 6 · Roles & Responsibilities
    "<h2>6 · Roles &amp; Responsibilities</h2>",
    "<p></p>",
    "<table><thead><tr><th>Role</th><th>Responsibilities</th></tr></thead><tbody>",
    "<tr><td>🎯 <strong>Data Owner</strong></td><td>Assign classification · Approve access · Define retention · Review annually</td></tr>",
    "<tr><td>🔧 <strong>Data Custodian</strong></td><td>Implement technical controls · Manage storage · Execute disposal</td></tr>",
    "<tr><td>👤 <strong>Data User</strong></td><td>Handle data per classification · Report misclassification · Complete training</td></tr>",
    "<tr><td>🛡️ <strong>Security Team</strong></td><td>Audit compliance · Monitor access · Investigate violations</td></tr>",
    "<tr><td>⚖️ <strong>Legal / Compliance</strong></td><td>Map regulatory requirements · Advise on cross-border transfers</td></tr>",
    "</tbody></table>",
    "<p></p>",

    // 7 · Implementation Checklist
    "<h2>7 · Implementation Checklist</h2>",
    "<ul data-type=\"taskList\">",
    "<li data-type=\"taskItem\" data-checked=\"false\">Complete a data inventory of all critical information assets</li>",
    "<li data-type=\"taskItem\" data-checked=\"false\">Assign a Data Owner to each identified asset</li>",
    "<li data-type=\"taskItem\" data-checked=\"false\">Classify all assets using the levels defined above</li>",
    "<li data-type=\"taskItem\" data-checked=\"false\">Apply labels to all documents, systems, and storage</li>",
    "<li data-type=\"taskItem\" data-checked=\"false\">Implement access controls per the handling matrix</li>",
    "<li data-type=\"taskItem\" data-checked=\"false\">Configure encryption for Confidential and Secret data</li>",
    "<li data-type=\"taskItem\" data-checked=\"false\">Train all employees on classification procedures</li>",
    "<li data-type=\"taskItem\" data-checked=\"false\">Schedule annual review of all classifications</li>",
    "</ul>",
    "<p></p>",
    "<hr>",
    "<p></p>",
    "<blockquote><p>📎 <strong>Related Documents:</strong> Information Security Policy · Acceptable Use Policy · Data Retention Schedule · Privacy Policy · Access Control Policy</p></blockquote>",
    "<p></p>",
    "<p><em>Proper data classification is the foundation of information security. When data is correctly classified, the right protections follow naturally.</em></p>",
  ].join(""),
};
