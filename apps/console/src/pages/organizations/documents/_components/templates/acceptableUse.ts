import { IconPageTextLine } from "@probo/ui";

import type { DocumentTemplate } from "./types";

export const acceptableUseTemplate: DocumentTemplate = {
  id: "acceptable-use-policy",
  icon: IconPageTextLine,
  name: "Acceptable Use",
  documentType: "POLICY",
  classification: "INTERNAL",
  title: "Acceptable Use Policy",
  content: [
    "<h1>📄 Acceptable Use Policy</h1>",
    "<p></p>",

    "<table><thead><tr><th>Property</th><th>Details</th></tr></thead><tbody>",
    "<tr><td>📋 <strong>Document Owner</strong></td><td>IT Department</td></tr>",
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
    "<p>This policy establishes clear rules for how <strong>[Organization Name]</strong> employees and authorized personnel may use company IT resources, protecting both the organization and its people.</p>",
    "<p></p>",
    "<blockquote><p>💡 <strong>Why this matters:</strong> Proper use of IT resources prevents data breaches, ensures compliance with regulations, and maintains productivity across the organization.</p></blockquote>",
    "<p></p>",

    // 2 · Scope
    "<h2>2 · Scope</h2>",
    "<p>This policy covers <strong>all users</strong> and <strong>all IT resources</strong>:</p>",
    "<p></p>",
    "<table><thead><tr><th>Users</th><th>Resources</th></tr></thead><tbody>",
    "<tr><td>Full-time &amp; part-time employees</td><td>Laptops, desktops, mobile devices</td></tr>",
    "<tr><td>Contractors &amp; consultants</td><td>Email &amp; collaboration tools</td></tr>",
    "<tr><td>Temporary staff &amp; interns</td><td>Cloud services &amp; SaaS platforms</td></tr>",
    "<tr><td>Third-party vendors with access</td><td>Network &amp; internet access</td></tr>",
    "</tbody></table>",
    "<p></p>",

    // 3 · Use Guidelines
    "<h2>3 · Use Guidelines</h2>",
    "<p></p>",

    "<h3>3.1 — ✅ Permitted Activities</h3>",
    "<p></p>",
    "<table><thead><tr><th>Activity</th><th>Status</th><th>Conditions</th></tr></thead><tbody>",
    "<tr><td>Business email &amp; calendaring</td><td>✅ Allowed</td><td>Use official accounts for all business communication</td></tr>",
    "<tr><td>Approved SaaS tools</td><td>✅ Allowed</td><td>Only tools listed in the approved software catalog</td></tr>",
    "<tr><td>Company cloud storage</td><td>✅ Allowed</td><td>Use company-managed drives — no personal cloud storage for work data</td></tr>",
    "<tr><td>Limited personal browsing</td><td>⚠️ Limited</td><td>During breaks only · Must not affect productivity or bandwidth</td></tr>",
    "<tr><td>Personal devices (BYOD)</td><td>⚠️ Limited</td><td>Only with MDM enrollment and IT approval</td></tr>",
    "</tbody></table>",
    "<p></p>",

    "<h3>3.2 — 🚫 Prohibited Activities</h3>",
    "<p>The following actions are <strong>strictly forbidden</strong> and may result in immediate disciplinary action:</p>",
    "<p></p>",
    "<ul>",
    "<li>🔴 Accessing systems, data, or accounts <strong>without authorization</strong></li>",
    "<li>🔴 Installing <strong>unauthorized, pirated, or unlicensed</strong> software</li>",
    "<li>🔴 Sharing passwords, credentials, or <strong>access tokens</strong> with anyone</li>",
    "<li>🔴 Disabling antivirus, firewall, or any <strong>security controls</strong></li>",
    "<li>🔴 Sending confidential data through <strong>personal email or messaging apps</strong></li>",
    "<li>🔴 Connecting <strong>unauthorized devices</strong> to the corporate network</li>",
    "<li>🔴 Using resources for <strong>illegal, harassing, or discriminatory</strong> purposes</li>",
    "</ul>",
    "<p></p>",

    // 4 · Security Requirements
    "<h2>4 · Security Requirements</h2>",
    "<p>Every user must follow these baseline security practices:</p>",
    "<p></p>",
    "<table><thead><tr><th>Requirement</th><th>Minimum Standard</th><th>Enforcement</th></tr></thead><tbody>",
    "<tr><td>🔑 <strong>Passwords</strong></td><td>12+ characters · Unique per service · Stored in approved password manager</td><td>Automated policy</td></tr>",
    "<tr><td>📱 <strong>MFA</strong></td><td>Enabled on all accounts that support it</td><td>Mandatory enrollment</td></tr>",
    "<tr><td>🖥️ <strong>Screen Lock</strong></td><td>Auto-lock after 5 minutes of inactivity</td><td>MDM enforced</td></tr>",
    "<tr><td>🔄 <strong>Software Updates</strong></td><td>Install security patches within 48 hours</td><td>Automated + audit</td></tr>",
    "<tr><td>🎣 <strong>Phishing</strong></td><td>Report to security@[domain] — do not click, reply, or forward</td><td>Quarterly training</td></tr>",
    "<tr><td>💾 <strong>Data Handling</strong></td><td>Follow Data Classification Procedure for all information</td><td>Spot audits</td></tr>",
    "</tbody></table>",
    "<p></p>",

    // 5 · Email & Communication Standards
    "<h2>5 · Email &amp; Communication Standards</h2>",
    "<ul>",
    "<li>Use <strong>company email</strong> for all business correspondence — never personal accounts</li>",
    "<li>Do not open attachments or click links from <strong>unknown or suspicious senders</strong></li>",
    "<li>Include the <strong>classification label</strong> in the subject line when sending sensitive documents</li>",
    "<li>Use <strong>encrypted channels</strong> (company Slack, Teams) for internal sensitive discussions</li>",
    "</ul>",
    "<p></p>",

    // 6 · Monitoring & Privacy
    "<h2>6 · Monitoring &amp; Privacy</h2>",
    "<blockquote><p>⚠️ <strong>Notice:</strong> The organization reserves the right to monitor use of IT resources in accordance with applicable laws. This includes email, internet traffic, and access logs. Users should have <strong>no expectation of privacy</strong> when using company IT systems.</p></blockquote>",
    "<p></p>",

    // 7 · Enforcement
    "<h2>7 · Enforcement</h2>",
    "<p>Violations will be handled according to severity:</p>",
    "<p></p>",
    "<table><thead><tr><th>Severity</th><th>Example</th><th>Action</th></tr></thead><tbody>",
    "<tr><td>⚠️ Minor</td><td>Excessive personal browsing</td><td>Verbal warning &amp; coaching</td></tr>",
    "<tr><td>🟠 Moderate</td><td>Installing unauthorized software</td><td>Written warning &amp; access review</td></tr>",
    "<tr><td>🔴 Serious</td><td>Sharing credentials, disabling security</td><td>Suspension of access &amp; HR review</td></tr>",
    "<tr><td>🚨 Critical</td><td>Intentional data breach, illegal activity</td><td>Immediate termination &amp; legal action</td></tr>",
    "</tbody></table>",
    "<p></p>",
    "<hr>",
    "<p></p>",
    "<blockquote><p>📎 <strong>Related Documents:</strong> Information Security Policy · Data Classification Procedure · Remote Work Policy · BYOD Policy</p></blockquote>",
    "<p></p>",
    "<p><em>By accessing company IT resources, you confirm that you have read, understood, and agree to comply with this Acceptable Use Policy.</em></p>",
  ].join(""),
};
