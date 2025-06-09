"use client";

import { Card } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import { HelpCircle, X, User } from "lucide-react";
import {
  graphql,
  PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
  useMutation,
  ConnectionHandler,
} from "react-relay";
import { Suspense, useEffect, useState, useCallback, useRef } from "react";
import type { VendorViewQuery as VendorViewQueryType } from "./__generated__/VendorViewQuery.graphql";
import type { VendorViewDeleteComplianceReportMutation as DeleteComplianceReportMutationType } from "./__generated__/VendorViewDeleteComplianceReportMutation.graphql";
import type { VendorViewUploadComplianceReportMutation as UploadComplianceReportMutationType } from "./__generated__/VendorViewUploadComplianceReportMutation.graphql";
import type { VendorViewUpdateVendorMutation, VendorCategory } from "./__generated__/VendorViewUpdateVendorMutation.graphql";
import type { VendorViewCreateRiskAssessmentMutation } from "./__generated__/VendorViewCreateRiskAssessmentMutation.graphql";
import type { VendorViewAssessVendorMutation } from "./__generated__/VendorViewAssessVendorMutation.graphql";
import type {
  BusinessImpact,
  DataSensitivity,
} from "./__generated__/VendorViewCreateRiskAssessmentMutation.graphql";
import type { VendorViewQuery$data } from "./__generated__/VendorViewQuery.graphql";
import { useParams } from "react-router";
import { cn } from "@/lib/utils";
import { PageTemplate } from "@/components/PageTemplate";
import { VendorViewSkeleton } from "./VendorPage";
import PeopleSelector from "@/components/PeopleSelector";
import { formatDistanceToNow, format, isPast } from "date-fns";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";

interface VendorCategoryOption {
  value: VendorCategory;
  label: string;
}

const VENDOR_CATEGORIES: VendorCategoryOption[] = [
  { value: "ANALYTICS", label: "Analytics" },
  { value: "CLOUD_MONITORING", label: "Cloud Monitoring" },
  { value: "CLOUD_PROVIDER", label: "Cloud Provider" },
  { value: "COLLABORATION", label: "Collaboration" },
  { value: "CUSTOMER_SUPPORT", label: "Customer Support" },
  { value: "DATA_STORAGE_AND_PROCESSING", label: "Data Storage and Processing" },
  { value: "DOCUMENT_MANAGEMENT", label: "Document Management" },
  { value: "EMPLOYEE_MANAGEMENT", label: "Employee Management" },
  { value: "ENGINEERING", label: "Engineering" },
  { value: "FINANCE", label: "Finance" },
  { value: "IDENTITY_PROVIDER", label: "Identity Provider" },
  { value: "IT", label: "IT" },
  { value: "MARKETING", label: "Marketing" },
  { value: "OFFICE_OPERATIONS", label: "Office Operations" },
  { value: "OTHER", label: "Other" },
  { value: "PASSWORD_MANAGEMENT", label: "Password Management" },
  { value: "PRODUCT_AND_DESIGN", label: "Product and Design" },
  { value: "PROFESSIONAL_SERVICES", label: "Professional Services" },
  { value: "RECRUITING", label: "Recruiting" },
  { value: "SALES", label: "Sales" },
  { value: "SECURITY", label: "Security" },
  { value: "VERSION_CONTROL", label: "Version Control" }
];

const vendorViewQuery = graphql`
  query VendorViewQuery($vendorId: ID!, $organizationId: ID!) {
    node(id: $vendorId) {
      ... on Vendor {
        id
        name
        description
        statusPageUrl
        termsOfServiceUrl
        privacyPolicyUrl
        serviceLevelAgreementUrl
        dataProcessingAgreementUrl
        businessAssociateAgreementUrl
        subprocessorsListUrl
        securityPageUrl
        trustPageUrl
        certifications
        headquarterAddress
        legalName
        websiteUrl
        category
        businessOwner {
          id
          fullName
        }
        securityOwner {
          id
          fullName
        }
        createdAt
        updatedAt
        complianceReports(first: 100)
          @connection(key: "VendorView_complianceReports") {
          edges {
            node {
              id
              reportName
              reportDate
              validUntil
              fileUrl
              fileSize
              createdAt
            }
          }
        }
        riskAssessments(first: 100)
          @connection(key: "VendorView_riskAssessments") {
          edges {
            node {
              id
              assessedAt
              expiresAt
              dataSensitivity
              businessImpact
              notes
              assessedBy {
                id
                fullName
              }
              createdAt
            }
          }
        }
      }
    }
    organization: node(id: $organizationId) {
      ...PeopleSelector_organization
    }
    viewer {
      user {
        people(organizationId: $organizationId) {
          id
        }
      }
    }
  }
`;

const updateVendorMutation = graphql`
  mutation VendorViewUpdateVendorMutation($input: UpdateVendorInput!) {
    updateVendor(input: $input) {
      vendor {
        id
        name
        description
        statusPageUrl
        termsOfServiceUrl
        privacyPolicyUrl
        serviceLevelAgreementUrl
        dataProcessingAgreementUrl
        businessAssociateAgreementUrl
        subprocessorsListUrl
        securityPageUrl
        trustPageUrl
        certifications
        headquarterAddress
        legalName
        websiteUrl
        category
        businessOwner {
          id
          fullName
        }
        securityOwner {
          id
          fullName
        }
        updatedAt
      }
    }
  }
`;

const deleteComplianceReportMutation = graphql`
  mutation VendorViewDeleteComplianceReportMutation(
    $input: DeleteVendorComplianceReportInput!
    $connections: [ID!]!
  ) {
    deleteVendorComplianceReport(input: $input) {
      deletedVendorComplianceReportId @deleteEdge(connections: $connections)
    }
  }
`;

const uploadComplianceReportMutation = graphql`
  mutation VendorViewUploadComplianceReportMutation(
    $input: UploadVendorComplianceReportInput!
    $connections: [ID!]!
  ) {
    uploadVendorComplianceReport(input: $input) {
      vendorComplianceReportEdge @appendEdge(connections: $connections) {
        node {
          id
          reportName
          reportDate
          validUntil
          fileUrl
          fileSize
          createdAt
        }
      }
    }
  }
`;

const createRiskAssessmentMutation = graphql`
  mutation VendorViewCreateRiskAssessmentMutation(
    $input: CreateVendorRiskAssessmentInput!
    $connections: [ID!]!
  ) {
    createVendorRiskAssessment(input: $input) {
      vendorRiskAssessmentEdge @prependEdge(connections: $connections) {
        node {
          id
          assessedAt
          expiresAt
          dataSensitivity
          businessImpact
          notes
          assessedBy {
            id
            fullName
          }
          createdAt
        }
      }
    }
  }
`;

const assessVendorMutation = graphql`
  mutation VendorViewAssessVendorMutation($input: AssessVendorInput!) {
    assessVendor(input: $input) {
      vendor {
        id
        name
        description
        statusPageUrl
        termsOfServiceUrl
        privacyPolicyUrl
        serviceLevelAgreementUrl
        dataProcessingAgreementUrl
        businessAssociateAgreementUrl
        subprocessorsListUrl
        securityPageUrl
        trustPageUrl
        certifications
        headquarterAddress
        legalName
        websiteUrl
        category
        businessOwner {
          id
          fullName
        }
        securityOwner {
          id
          fullName
        }
        updatedAt
      }
    }
  }
`;

// Format date for input field (YYYY-MM-DDTHH:mm)
function formatDateForInput(date: string | null | undefined): string {
  if (!date) return "";
  return new Date(date).toISOString().slice(0, 16);
}

// Format date for API (2006-01-02T15:04:05.999999999Z07:00)
function formatDateForAPI(dateStr: string): string {
  if (!dateStr) return "";
  const date = new Date(dateStr);
  return date.toISOString();
}

interface ComplianceReport {
  id: string;
  reportName: string;
  reportDate: string;
  validUntil: string | null | undefined;
  fileUrl: string;
  fileSize: number;
  createdAt: string;
}

function ComplianceReportsTable({
  reports,
  onDelete,
  onUpload,
}: {
  reports: ComplianceReport[];
  onDelete: (id: string) => void;
  onUpload: (event: React.ChangeEvent<HTMLInputElement>) => void;
}) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [showDropdown, setShowDropdown] = useState<string | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setShowDropdown(null);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  // Format file size to human-readable format
  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
    return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
  };

  // Format date to match Figma design
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("en-US", {
      weekday: "short",
      day: "numeric",
      month: "short",
      year: "numeric",
    });
  };

  return (
    <div className="space-y-4">
      <div className="rounded-xl border border-[#ECEFEC] bg-white overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-[rgba(2,42,2,0.08)]">
              <th className="px-4 py-3 text-left">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold text-[#818780]">
                    Report name
                  </span>
                </div>
              </th>
              <th className="px-4 py-3 text-left">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold text-[#818780]">
                    Report date
                  </span>
                </div>
              </th>
              <th className="px-4 py-3 text-left">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold text-[#818780]">
                    Valid until
                  </span>
                </div>
              </th>
              <th className="px-4 py-3 text-right w-16"></th>
            </tr>
          </thead>
          <tbody>
            {reports.map((report) => (
              <tr
                key={report.id}
                className="border-b border-[rgba(2,42,2,0.08)]"
              >
                <td className="px-4 py-3">
                  <div className="flex flex-col justify-center">
                    <a
                      href={report.fileUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm font-normal text-[#141E12] hover:underline"
                    >
                      {report.reportName}
                    </a>
                    <span className="text-xs text-[#818780]">
                      {formatFileSize(report.fileSize)}
                    </span>
                  </div>
                </td>
                <td className="px-4 py-3">
                  <span className="text-sm font-normal text-[#141E12]">
                    {formatDate(report.reportDate)}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className="text-sm font-normal text-[#141E12]">
                    {report.validUntil ? formatDate(report.validUntil) : "N/A"}
                  </span>
                </td>
                <td className="text-right pr-6 py-3">
                  <div
                    className="relative"
                    ref={showDropdown === report.id ? dropdownRef : null}
                  >
                    <button
                      className="rounded-full w-8 h-8 flex items-center justify-center hover:bg-[rgba(2,42,2,0.03)]"
                      onClick={() =>
                        setShowDropdown(
                          showDropdown === report.id ? null : report.id,
                        )
                      }
                    >
                      <svg
                        width="16"
                        height="16"
                        viewBox="0 0 16 16"
                        fill="none"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path
                          d="M8 9.5C8.82843 9.5 9.5 8.82843 9.5 8C9.5 7.17157 8.82843 6.5 8 6.5C7.17157 6.5 6.5 7.17157 6.5 8C6.5 8.82843 7.17157 9.5 8 9.5Z"
                          fill="#141E12"
                        />
                        <path
                          d="M3 9.5C3.82843 9.5 4.5 8.82843 4.5 8C4.5 7.17157 3.82843 6.5 3 6.5C2.17157 6.5 1.5 7.17157 1.5 8C1.5 8.82843 2.17157 9.5 3 9.5Z"
                          fill="#141E12"
                        />
                        <path
                          d="M13 9.5C13.8284 9.5 14.5 8.82843 14.5 8C14.5 7.17157 13.8284 6.5 13 6.5C12.1716 6.5 11.5 7.17157 11.5 8C11.5 8.82843 12.1716 9.5 13 9.5Z"
                          fill="#141E12"
                        />
                      </svg>
                    </button>
                    {showDropdown === report.id && (
                      <div
                        className="absolute right-0 mt-1 bg-white rounded-md shadow-lg border border-[#ECEFEC] z-10"
                        style={{ bottom: "100%", marginBottom: "5px" }}
                      >
                        <button
                          className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50"
                          onClick={() => {
                            onDelete(report.id);
                            setShowDropdown(null);
                          }}
                        >
                          Delete
                        </button>
                      </div>
                    )}
                  </div>
                </td>
              </tr>
            ))}
            {reports.length === 0 && (
              <tr>
                <td
                  colSpan={4}
                  className="text-center py-6 text-sm text-[#818780]"
                >
                  No compliance reports uploaded yet
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <div className="mt-4">
        <input
          type="file"
          ref={fileInputRef}
          className="hidden"
          onChange={onUpload}
          accept=".pdf"
        />
        <Button
          onClick={() => fileInputRef.current?.click()}
          className="rounded-full px-4 py-2 bg-[#054D05] text-white hover:bg-[#054D05]/90"
          size="sm"
        >
          Upload New Report
        </Button>
        <p className="mt-2 text-sm text-[#6B716A]">
          Only PDF files up to 10MB are allowed
        </p>
      </div>
    </div>
  );
}

function CertificationsView({
  certifications,
  onAdd,
  onRemove,
}: {
  certifications: readonly string[];
  onAdd: (tag: string) => void;
  onRemove: (tag: string) => void;
}) {
  // Common security standards
  const securityStandards = [
    "SOC 2",
    "ISO 27001",
    "HITRUST",
    "NIST",
    "SOC 2 Type 2",
    "SOC 2 Type 1",
  ];

  // Regulatory and legal certifications
  const regulatoryLegal = ["HIPAA", "FERPA", "FISMA", "PIPEDA", "GDPR", "CCPA"];

  // Industry-specific certifications
  const industrySpecific = ["FinTech", "MPAA", "GSMA"];

  // International & government certifications
  const internationalGov = ["FedRAMP", "ENS High", "IRAP", "CJIS"];

  // State for new certification input
  const [showAddModal, setShowAddModal] = useState(false);
  const [newCertification, setNewCertification] = useState("");

  // Handle adding a new custom certification
  const handleAddCertification = () => {
    if (newCertification.trim()) {
      onAdd(newCertification.trim());
      setNewCertification("");
      setShowAddModal(false);
    }
  };

  // Get certifications by category
  const securityCerts = certifications.filter((cert) =>
    securityStandards.includes(cert),
  );
  const regulatoryCerts = certifications.filter((cert) =>
    regulatoryLegal.includes(cert),
  );
  const industryCerts = certifications.filter((cert) =>
    industrySpecific.includes(cert),
  );
  const internationalCerts = certifications.filter((cert) =>
    internationalGov.includes(cert),
  );
  const customCerts = certifications.filter(
    (cert) =>
      ![
        ...securityStandards,
        ...regulatoryLegal,
        ...industrySpecific,
        ...internationalGov,
      ].includes(cert),
  );

  return (
    <div className="rounded-xl border border-[#ECEFEC] bg-white overflow-hidden">
      <div className="flex items-center justify-between px-5 py-3 border-b border-[rgba(2,42,2,0.08)]">
        <h3 className="text-base font-medium text-[#141E12]">Certifications</h3>
        <Button
          onClick={() => setShowAddModal(true)}
          className="rounded-full bg-[rgba(0,39,0,0.05)] text-[#141E12] hover:bg-[rgba(0,39,0,0.08)] h-8 px-3"
          size="sm"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 16 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            className="mr-1.5"
          >
            <path
              d="M8 3.5V12.5"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path
              d="M3.5 8H12.5"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          Add certification
        </Button>
      </div>

      <div className="p-5 space-y-6">
        {/* Security Standards */}
        {securityCerts.length > 0 && (
          <div className="space-y-3">
            <h4 className="text-sm font-medium text-[#141E12]">
              Security Standards
            </h4>
            <div className="flex flex-wrap gap-2">
              {securityCerts.map((cert) => (
                <div
                  key={cert}
                  className="flex items-center gap-1 rounded bg-[rgba(5,77,5,0.03)] px-2 py-1.5 text-sm font-medium text-[#6B716A]"
                >
                  <span>{cert}</span>
                  <button
                    onClick={() => onRemove(cert)}
                    className="text-[#6B716A] hover:text-[#141E12]"
                  >
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 16 16"
                      fill="none"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <path
                        d="M12 4L4 12"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                      <path
                        d="M4 4L12 12"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Regulatory & Legal */}
        {regulatoryCerts.length > 0 && (
          <div className="space-y-3">
            <h4 className="text-sm font-medium text-[#141E12]">
              Regulatory & Legal
            </h4>
            <div className="flex flex-wrap gap-2">
              {regulatoryCerts.map((cert) => (
                <div
                  key={cert}
                  className="flex items-center gap-1 rounded bg-[rgba(5,77,5,0.03)] px-2 py-1.5 text-sm font-medium text-[#6B716A]"
                >
                  <span>{cert}</span>
                  <button
                    onClick={() => onRemove(cert)}
                    className="text-[#6B716A] hover:text-[#141E12]"
                  >
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 16 16"
                      fill="none"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <path
                        d="M12 4L4 12"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                      <path
                        d="M4 4L12 12"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Industry-Specific */}
        {industryCerts.length > 0 && (
          <div className="space-y-3">
            <h4 className="text-sm font-medium text-[#141E12]">
              Industry-Specific
            </h4>
            <div className="flex flex-wrap gap-2">
              {industryCerts.map((cert) => (
                <div
                  key={cert}
                  className="flex items-center gap-1 rounded bg-[rgba(5,77,5,0.03)] px-2 py-1.5 text-sm font-medium text-[#6B716A]"
                >
                  <span>{cert}</span>
                  <button
                    onClick={() => onRemove(cert)}
                    className="text-[#6B716A] hover:text-[#141E12]"
                  >
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 16 16"
                      fill="none"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <path
                        d="M12 4L4 12"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                      <path
                        d="M4 4L12 12"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* International & Government */}
        {internationalCerts.length > 0 && (
          <div className="space-y-3">
            <h4 className="text-sm font-medium text-[#141E12]">
              International & Government
            </h4>
            <div className="flex flex-wrap gap-2">
              {internationalCerts.map((cert) => (
                <div
                  key={cert}
                  className="flex items-center gap-1 rounded bg-[rgba(5,77,5,0.03)] px-2 py-1.5 text-sm font-medium text-[#6B716A]"
                >
                  <span>{cert}</span>
                  <button
                    onClick={() => onRemove(cert)}
                    className="text-[#6B716A] hover:text-[#141E12]"
                  >
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 16 16"
                      fill="none"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <path
                        d="M12 4L4 12"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                      <path
                        d="M4 4L12 12"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Custom certifications */}
        {customCerts.length > 0 && (
          <div className="space-y-3">
            <h4 className="text-sm font-medium text-[#141E12]">
              Custom Certifications
            </h4>
            <div className="flex flex-wrap gap-2">
              {customCerts.map((cert) => (
                <div
                  key={cert}
                  className="flex items-center gap-1 rounded bg-[rgba(5,77,5,0.03)] px-2 py-1.5 text-sm font-medium text-[#6B716A]"
                >
                  <span>{cert}</span>
                  <button
                    onClick={() => onRemove(cert)}
                    className="text-[#6B716A] hover:text-[#141E12]"
                  >
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 16 16"
                      fill="none"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <path
                        d="M12 4L4 12"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                      <path
                        d="M4 4L12 12"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Add Certification Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md">
            <h3 className="text-lg font-medium mb-4">
              Add Custom Certification
            </h3>
            <div className="space-y-4">
              <div>
                <Label htmlFor="certification-name">Certification Name</Label>
                <Input
                  id="certification-name"
                  value={newCertification}
                  onChange={(e) => setNewCertification(e.target.value)}
                  placeholder="Enter certification name"
                  className="mt-1"
                />
              </div>
              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  onClick={() => setShowAddModal(false)}
                >
                  Cancel
                </Button>
                <Button onClick={handleAddCertification}>Add</Button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function RiskAssessmentsTable({
  assessments,
  onCreateAssessment,
}: {
  assessments: NonNullable<
    NonNullable<VendorViewQuery$data["node"]>["riskAssessments"]
  >["edges"];
  onCreateAssessment: () => void;
}) {
  const [showDropdown, setShowDropdown] = useState<string | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const [selectedAssessment, setSelectedAssessment] = useState<any>(null);
  const [showDetailsModal, setShowDetailsModal] = useState(false);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setShowDropdown(null);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  // Format date to relative time using date-fns
  const formatRelativeDate = (dateString: string) => {
    const date = new Date(dateString);
    return formatDistanceToNow(date, { addSuffix: true });
  };

  // Format date for tooltip
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return format(date, "EEE, d MMM yyyy");
  };

  const isExpired = (expiresAt: string) => isPast(new Date(expiresAt));

  // Handle click on view details
  const handleViewDetails = (assessment: any) => {
    setSelectedAssessment(assessment);
    setShowDetailsModal(true);
    setShowDropdown(null);
  };

  return (
    <div className="space-y-4">
      <div className="rounded-xl border border-[#ECEFEC] bg-white overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-[rgba(2,42,2,0.08)]">
              <th className="px-4 py-3 text-left">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold text-[#818780]">
                    Assessed By
                  </span>
                </div>
              </th>
              <th className="px-4 py-3 text-left">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold text-[#818780]">
                    Expires
                  </span>
                </div>
              </th>
              <th className="px-4 py-3 text-left">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold text-[#818780]">
                    Data Sensitivity
                  </span>
                </div>
              </th>
              <th className="px-4 py-3 text-left">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold text-[#818780]">
                    Business Impact
                  </span>
                </div>
              </th>
              <th className="px-4 py-3 text-right w-16"></th>
            </tr>
          </thead>
          <tbody>
            {assessments?.map((edge) => {
              const assessment = edge?.node;
              const expired = isExpired(assessment?.expiresAt || "");
              return (
                <tr
                  key={assessment?.id}
                  className={cn(
                    "border-b border-[rgba(2,42,2,0.08)]",
                    expired && "bg-[#F7F7F7] text-[#A0A5A0]",
                  )}
                >
                  <td className="px-4 py-3">
                    <span
                      className={cn(
                        "text-sm font-normal",
                        expired ? "text-[#A0A5A0]" : "text-[#141E12]",
                      )}
                    >
                      {assessment?.assessedBy?.fullName || "N/A"} (
                      <span
                        className={cn(
                          "text-sm font-normal",
                          expired ? "text-[#A0A5A0]" : "text-[#141E12]",
                        )}
                        title={formatDate(assessment?.assessedAt || "")}
                      >
                        {formatRelativeDate(assessment?.assessedAt || "")}
                      </span>
                      )
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center">
                      <span
                        className={cn(
                          "text-sm font-normal",
                          expired ? "text-[#A0A5A0]" : "text-[#141E12]",
                        )}
                        title={formatDate(assessment?.expiresAt || "")}
                      >
                        {formatRelativeDate(assessment?.expiresAt || "")}
                      </span>
                      {expired && (
                        <span className="ml-2 text-xs py-0.5 px-1.5 bg-[#F0F0F0] text-[#818780] rounded">
                          Expired
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={cn(
                        "text-sm font-normal",
                        expired ? "text-[#A0A5A0]" : "text-[#141E12]",
                      )}
                    >
                      {assessment?.dataSensitivity.charAt(0) +
                        assessment?.dataSensitivity.slice(1).toLowerCase()}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={cn(
                        "text-sm font-normal",
                        expired ? "text-[#A0A5A0]" : "text-[#141E12]",
                      )}
                    >
                      {assessment?.businessImpact.charAt(0) +
                        assessment?.businessImpact.slice(1).toLowerCase()}
                    </span>
                  </td>
                  <td className="text-right pr-6 py-3">
                    <div className="relative" ref={showDropdown === assessment?.id ? dropdownRef : null}>
                      <button
                        className="rounded-full w-8 h-8 flex items-center justify-center hover:bg-[rgba(2,42,2,0.03)]"
                        onClick={() =>
                          setShowDropdown(
                            showDropdown === assessment?.id ? null : assessment?.id
                          )
                        }
                      >
                        <svg
                          width="16"
                          height="16"
                          viewBox="0 0 16 16"
                          fill="none"
                          xmlns="http://www.w3.org/2000/svg"
                        >
                          <path
                            d="M8 9.5C8.82843 9.5 9.5 8.82843 9.5 8C9.5 7.17157 8.82843 6.5 8 6.5C7.17157 6.5 6.5 7.17157 6.5 8C6.5 8.82843 7.17157 9.5 8 9.5Z"
                            fill={expired ? "#A0A5A0" : "#141E12"}
                          />
                          <path
                            d="M3 9.5C3.82843 9.5 4.5 8.82843 4.5 8C4.5 7.17157 3.82843 6.5 3 6.5C2.17157 6.5 1.5 7.17157 1.5 8C1.5 8.82843 2.17157 9.5 3 9.5Z"
                            fill={expired ? "#A0A5A0" : "#141E12"}
                          />
                          <path
                            d="M13 9.5C13.8284 9.5 14.5 8.82843 14.5 8C14.5 7.17157 13.8284 6.5 13 6.5C12.1716 6.5 11.5 7.17157 11.5 8C11.5 8.82843 12.1716 9.5 13 9.5Z"
                            fill={expired ? "#A0A5A0" : "#141E12"}
                          />
                        </svg>
                      </button>
                      {showDropdown === assessment?.id && (
                        <div className="absolute right-0 bottom-full mb-2 bg-white rounded-md shadow-lg border border-[#ECEFEC] z-10 w-32">
                          <button
                            className="w-full text-left px-4 py-2 text-sm hover:bg-[rgba(2,42,2,0.03)] whitespace-nowrap"
                            onClick={() => handleViewDetails(assessment)}
                          >
                            View details
                          </button>
                        </div>
                      )}
                    </div>
                  </td>
                </tr>
              );
            })}
            {assessments?.length === 0 && (
              <tr>
                <td
                  colSpan={7}
                  className="text-center py-6 text-sm text-[#818780]"
                >
                  No risk assessments created yet
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <div className="mt-4">
        <Button
          onClick={onCreateAssessment}
          className="rounded-full px-4 py-2 bg-[#054D05] text-white hover:bg-[#054D05]/90"
          size="sm"
        >
          Create Risk Assessment
        </Button>
        <p className="mt-2 text-sm text-[#6B716A]">
          Create a new risk assessment for this vendor
        </p>
      </div>

      {/* Risk Assessment Details Modal */}
      {showDetailsModal && selectedAssessment && (
        <RiskAssessmentDetailsModal
          assessment={selectedAssessment}
          onClose={() => setShowDetailsModal(false)}
        />
      )}
    </div>
  );
}

function RiskAssessmentDetailsModal({
  assessment,
  onClose,
}: {
  assessment: any;
  onClose: () => void;
}) {
  // Format date for display
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return format(date, "MMMM d, yyyy");
  };

  // Data sensitivity descriptions
  const dataSensitivityDescriptions: Record<string, string> = {
    NONE: "No sensitive data",
    LOW: "Public or non-sensitive data",
    MEDIUM: "Internal/restricted data",
    HIGH: "Confidential data",
    CRITICAL: "Regulated/PII/financial data",
  };

  // Business impact descriptions
  const businessImpactDescriptions: Record<string, string> = {
    LOW: "Minimal impact on business",
    MEDIUM: "Moderate impact on business",
    HIGH: "Significant business impact",
    CRITICAL: "Critical to business operations",
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-medium">Risk Assessment Details</h3>
          <Button
            variant="ghost"
            size="icon"
            onClick={onClose}
            className="h-8 w-8 rounded-full"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>

        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <h4 className="text-sm font-medium text-[#6B716A] mb-1">Assessed By</h4>
              <div className="flex items-center space-x-2">
                <User className="h-5 w-5 text-[#6B716A]" />
                <p className="text-sm font-medium">{assessment.assessedBy?.fullName || "N/A"}</p>
              </div>
              <p className="text-xs text-[#818780] mt-1">
                {formatDate(assessment.assessedAt || "")}
              </p>
            </div>

            <div>
              <h4 className="text-sm font-medium text-[#6B716A] mb-1">Valid Until</h4>
              <p className="text-sm font-medium">
                {formatDate(assessment.expiresAt || "")}
              </p>
              {isPast(new Date(assessment.expiresAt || "")) && (
                <p className="text-xs text-red-500 mt-1">
                  Assessment has expired
                </p>
              )}
            </div>
          </div>

          <div>
            <h4 className="text-sm font-medium text-[#6B716A] mb-2">Data Sensitivity</h4>
            <div className="p-3 bg-[rgba(5,77,5,0.03)] rounded-md">
              <p className="text-sm font-medium text-[#141E12]">
                {assessment.dataSensitivity.charAt(0) + assessment.dataSensitivity.slice(1).toLowerCase()}
              </p>
              <p className="text-sm text-[#6B716A] mt-1">
                {dataSensitivityDescriptions[assessment.dataSensitivity] || ""}
              </p>
            </div>
          </div>

          <div>
            <h4 className="text-sm font-medium text-[#6B716A] mb-2">Business Impact</h4>
            <div className="p-3 bg-[rgba(5,77,5,0.03)] rounded-md">
              <p className="text-sm font-medium text-[#141E12]">
                {assessment.businessImpact.charAt(0) + assessment.businessImpact.slice(1).toLowerCase()}
              </p>
              <p className="text-sm text-[#6B716A] mt-1">
                {businessImpactDescriptions[assessment.businessImpact] || ""}
              </p>
            </div>
          </div>

          {assessment.notes && (
            <div>
              <h4 className="text-sm font-medium text-[#6B716A] mb-2">Notes</h4>
              <div className="p-3 bg-[rgba(5,77,5,0.03)] rounded-md">
                <p className="text-sm text-[#141E12] whitespace-pre-wrap">{assessment.notes}</p>
              </div>
            </div>
          )}

          <div className="flex justify-end mt-4">
            <Button
              onClick={onClose}
              className="bg-[#054D05] text-white hover:bg-[#054D05]/90"
            >
              Close
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

function RiskAssessmentModal({
  isOpen,
  onClose,
  onSubmit,
  businessOwnerId,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: {
    dataSensitivity: DataSensitivity;
    businessImpact: BusinessImpact;
    notes: string;
  }) => void;
  businessOwnerId: string | null;
}) {
  const [dataSensitivity, setDataSensitivity] =
    useState<DataSensitivity>("LOW");
  const [businessImpact, setBusinessImpact] = useState<BusinessImpact>("LOW");
  const [notes, setNotes] = useState("");

  // Reset form when modal opens
  useEffect(() => {
    if (isOpen) {
      setDataSensitivity("LOW");
      setBusinessImpact("LOW");
      setNotes("");
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const handleSubmit = () => {
    onSubmit({
      dataSensitivity,
      businessImpact,
      notes,
    });
  };

  // Data sensitivity descriptions from vendor_risk_assessment.go
  const dataSensitivityOptions = [
    { value: "NONE", label: "None", description: "No sensitive data" },
    { value: "LOW", label: "Low", description: "Public or non-sensitive data" },
    {
      value: "MEDIUM",
      label: "Medium",
      description: "Internal/restricted data",
    },
    { value: "HIGH", label: "High", description: "Confidential data" },
    {
      value: "CRITICAL",
      label: "Critical",
      description: "Regulated/PII/financial data",
    },
  ];

  // Business impact descriptions from vendor_risk_assessment.go
  const businessImpactOptions = [
    { value: "LOW", label: "Low", description: "Minimal impact on business" },
    {
      value: "MEDIUM",
      label: "Medium",
      description: "Moderate impact on business",
    },
    {
      value: "HIGH",
      label: "High",
      description: "Significant business impact",
    },
    {
      value: "CRITICAL",
      label: "Critical",
      description: "Critical to business operations",
    },
  ];

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-6 w-full max-w-md max-h-[90vh] overflow-y-auto">
        <h3 className="text-lg font-medium mb-4">Create Risk Assessment</h3>

        <div className="space-y-6">
          <div>
            <Label htmlFor="data-sensitivity" className="mb-2 block">
              Data Sensitivity
            </Label>
            <select
              id="data-sensitivity"
              value={dataSensitivity}
              onChange={(e) =>
                setDataSensitivity(e.target.value as DataSensitivity)
              }
              className="w-full rounded-md border border-[#ECEFEC] px-3 py-2 focus:border-[#054D05] focus:ring-[#054D05]"
            >
              {dataSensitivityOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            <div className="mt-2 p-3 bg-[rgba(5,77,5,0.03)] rounded-md">
              <p className="text-sm text-[#141E12] font-medium mb-1">
                {
                  dataSensitivityOptions.find(
                    (o) => o.value === dataSensitivity,
                  )?.label
                }
                :
              </p>
              <p className="text-sm text-[#6B716A]">
                {
                  dataSensitivityOptions.find(
                    (o) => o.value === dataSensitivity,
                  )?.description
                }
              </p>
              <p className="mt-2 text-xs text-[#6B716A]">
                Select the level of sensitivity for the data this vendor
                processes or stores.
              </p>
            </div>
          </div>

          <div>
            <Label htmlFor="business-impact" className="mb-2 block">
              Business Impact
            </Label>
            <select
              id="business-impact"
              value={businessImpact}
              onChange={(e) =>
                setBusinessImpact(e.target.value as BusinessImpact)
              }
              className="w-full rounded-md border border-[#ECEFEC] px-3 py-2 focus:border-[#054D05] focus:ring-[#054D05]"
            >
              {businessImpactOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            <div className="mt-2 p-3 bg-[rgba(5,77,5,0.03)] rounded-md">
              <p className="text-sm text-[#141E12] font-medium mb-1">
                {
                  businessImpactOptions.find((o) => o.value === businessImpact)
                    ?.label
                }
                :
              </p>
              <p className="text-sm text-[#6B716A]">
                {
                  businessImpactOptions.find((o) => o.value === businessImpact)
                    ?.description
                }
              </p>
              <p className="mt-2 text-xs text-[#6B716A]">
                Select the impact on your business if this vendor experiences an
                outage or data breach.
              </p>
            </div>
          </div>

          <div>
            <Label htmlFor="notes" className="mb-2 block">
              Notes
            </Label>
            <textarea
              id="notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              className="w-full rounded-md border border-[#ECEFEC] px-3 py-2 min-h-[80px]"
              placeholder="Add any additional notes about this risk assessment"
            />
            <p className="mt-2 text-xs text-[#6B716A]">
              Add any context or details about this risk assessment that might
              be helpful for future reference.
            </p>
          </div>

          <div className="flex justify-end gap-2 pt-4">
            <Button
              variant="outline"
              onClick={onClose}
              className="border-[#ECEFEC] text-[#141E12]"
            >
              Cancel
            </Button>
            <Button
              onClick={handleSubmit}
              className="bg-[#054D05] text-white hover:bg-[#054D05]/90"
            >
              Create Assessment
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

function VendorViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<VendorViewQueryType>;
}) {
  const { organizationId } = useParams();
  const data = usePreloadedQuery(vendorViewQuery, queryRef);
  const [activeTab, setActiveTab] = useState<
    "overview" | "certifications" | "complianceReports" | "riskAssessments"
  >("overview");
  const [editedFields, setEditedFields] = useState<Set<string>>(new Set());
  const [showRiskAssessmentModal, setShowRiskAssessmentModal] = useState(false);
  const [formData, setFormData] = useState({
    name: data.node.name || "",
    description: data.node.description || "",
    statusPageUrl: data.node.statusPageUrl || "",
    termsOfServiceUrl: data.node.termsOfServiceUrl || "",
    privacyPolicyUrl: data.node.privacyPolicyUrl || "",
    serviceLevelAgreementUrl: data.node.serviceLevelAgreementUrl || "",
    dataProcessingAgreementUrl: data.node.dataProcessingAgreementUrl || "",
    businessAssociateAgreementUrl: data.node.businessAssociateAgreementUrl || "",
    subprocessorsListUrl: data.node.subprocessorsListUrl || "",
    securityPageUrl: data.node.securityPageUrl || "",
    trustPageUrl: data.node.trustPageUrl || "",
    certifications: data.node.certifications || [],
    headquarterAddress: data.node.headquarterAddress || "",
    legalName: data.node.legalName || "",
    websiteUrl: data.node.websiteUrl || "",
    businessOwnerId: data.node.businessOwner?.id || null,
    securityOwnerId: data.node.securityOwner?.id || null,
    category: data.node.category || null,
  });
  const [updateVendor] =
    useMutation<VendorViewUpdateVendorMutation>(updateVendorMutation);
  const [deleteVendorComplianceReport] =
    useMutation<DeleteComplianceReportMutationType>(
      deleteComplianceReportMutation,
    );
  const [uploadVendorComplianceReport] =
    useMutation<UploadComplianceReportMutationType>(
      uploadComplianceReportMutation,
    );
  const [createRiskAssessment] =
    useMutation<VendorViewCreateRiskAssessmentMutation>(
      createRiskAssessmentMutation,
    );
  const [, loadQuery] = useQueryLoader<VendorViewQueryType>(vendorViewQuery);
  const { toast } = useToast();
  const [isAssessDialogOpen, setIsAssessDialogOpen] = useState(false);

  const [websiteUrl, setWebsiteUrl] = useState("");

  const [commitAssessVendor, isInFlight] = useMutation<VendorViewAssessVendorMutation>(
    assessVendorMutation
  );

  const hasChanges = editedFields.size > 0;

  const handleSave = useCallback(() => {
    const formattedData = {
      ...formData,
      businessOwnerId: formData.businessOwnerId || undefined,
      securityOwnerId: formData.securityOwnerId || undefined,
      category: formData.category as VendorCategory | null,
    };

    updateVendor({
      variables: {
        input: {
          id: data.node.id!,
          ...formattedData,
        },
      },
      onCompleted: () => {
        toast({
          title: "Success",
          description: "Changes saved successfully",
          variant: "default",
        });
        setEditedFields(new Set());
        loadQuery({ vendorId: data.node.id!, organizationId: organizationId! });
      },
      onError: (error) => {
        if (error.message?.includes("concurrent modification")) {
          toast({
            title: "Error",
            description:
              "Someone else modified this vendor. Reloading latest data.",
            variant: "destructive",
          });

          loadQuery({
            vendorId: data.node.id!,
            organizationId: organizationId!,
          });
        } else {
          toast({
            title: "Error",
            description: error.message || "Failed to save changes",
            variant: "destructive",
          });
        }
      },
    });
  }, [updateVendor, data.node.id, formData, loadQuery, toast, organizationId]);

  const handleFieldChange = (field: keyof typeof formData, value: unknown) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
    setEditedFields((prev) => new Set(prev).add(field));
  };

  // Auto-save handler specifically for certifications
  const handleCertificationChange = useCallback(
    (newCertifications: string[]) => {
      // Update local state
      setFormData((prev) => ({
        ...prev,
        certifications: newCertifications,
      }));

      // Auto-save the changes immediately
      const formattedData = {
        id: data.node.id!,
        certifications: newCertifications,
      };

      updateVendor({
        variables: {
          input: formattedData,
        },
        onCompleted: () => {
          toast({
            title: "Success",
            description: "Certifications updated",
            variant: "default",
          });
          loadQuery({
            vendorId: data.node.id!,
            organizationId: organizationId!,
          });
        },
        onError: (error) => {
          toast({
            title: "Error",
            description: error.message || "Failed to update certifications",
            variant: "destructive",
          });
        },
      });
    },
    [updateVendor, data.node.id, loadQuery, toast, organizationId],
  );

  const handleCancel = () => {
    setFormData({
      name: data.node.name || "",
      description: data.node.description || "",
      statusPageUrl: data.node.statusPageUrl || "",
      termsOfServiceUrl: data.node.termsOfServiceUrl || "",
      privacyPolicyUrl: data.node.privacyPolicyUrl || "",
      serviceLevelAgreementUrl: data.node.serviceLevelAgreementUrl || "",
      dataProcessingAgreementUrl: data.node.dataProcessingAgreementUrl || "",
      businessAssociateAgreementUrl: data.node.businessAssociateAgreementUrl || "",
      subprocessorsListUrl: data.node.subprocessorsListUrl || "",
      securityPageUrl: data.node.securityPageUrl || "",
      trustPageUrl: data.node.trustPageUrl || "",
      certifications: data.node.certifications || [],
      headquarterAddress: data.node.headquarterAddress || "",
      legalName: data.node.legalName || "",
      websiteUrl: data.node.websiteUrl || "",
      businessOwnerId: data.node.businessOwner?.id || null,
      securityOwnerId: data.node.securityOwner?.id || null,
      category: data.node.category || null,
    });
    setEditedFields(new Set());
  };

  const handleDeleteReport = useCallback(
    (reportId: string) => {
      deleteVendorComplianceReport({
        variables: {
          connections: [
            ConnectionHandler.getConnectionID(
              data.node.id!,
              "VendorView_complianceReports",
            ),
          ],
          input: {
            reportId,
          },
        },
        onCompleted: () => {
          toast({
            title: "Success",
            description: "Compliance report deleted successfully",
            variant: "default",
          });
          loadQuery({
            vendorId: data.node.id!,
            organizationId: organizationId!,
          });
        },
        onError: (error) => {
          toast({
            title: "Error",
            description: error.message || "Failed to delete compliance report",
            variant: "destructive",
          });
        },
      });
    },
    [
      deleteVendorComplianceReport,
      data.node.id,
      loadQuery,
      toast,
      organizationId,
    ],
  );

  const handleUploadReport = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const file = event.target.files?.[0];
      if (!file) return;

      // Validate file type
      if (file.type !== "application/pdf") {
        toast({
          title: "Error",
          description: "Only PDF files are allowed",
          variant: "destructive",
        });
        return;
      }

      // Validate file size (max 10MB)
      if (file.size > 10 * 1024 * 1024) {
        toast({
          title: "Error",
          description: "File size must be less than 10MB",
          variant: "destructive",
        });
        return;
      }

      const reader = new FileReader();
      reader.onload = () => {
        const reportDate = new Date().toISOString();

        uploadVendorComplianceReport({
          variables: {
            connections: [
              ConnectionHandler.getConnectionID(
                data.node.id!,
                "VendorView_complianceReports",
              ),
            ],
            input: {
              vendorId: data.node.id!,
              reportDate,
              reportName: file.name,
              file: null,
            },
          },
          uploadables: {
            "input.file": file,
          },
          onCompleted: () => {
            toast({
              title: "Success",
              description: "Compliance report uploaded successfully",
              variant: "default",
            });
            loadQuery({
              vendorId: data.node.id!,
              organizationId: organizationId!,
            });
          },
          onError: (error) => {
            toast({
              title: "Error",
              description:
                error.message || "Failed to upload compliance report",
              variant: "destructive",
            });
          },
        });
      };
      reader.readAsDataURL(file);
    },
    [
      uploadVendorComplianceReport,
      data.node.id,
      loadQuery,
      toast,
      organizationId,
    ],
  );

  const handleCreateRiskAssessment = useCallback(() => {
    setShowRiskAssessmentModal(true);
  }, []);

  const handleRiskAssessmentSubmit = useCallback(
    (formValues: {
      dataSensitivity: DataSensitivity;
      businessImpact: BusinessImpact;
      notes: string;
    }) => {
      const nextYear = new Date();
      nextYear.setFullYear(nextYear.getFullYear() + 1);

      createRiskAssessment({
        variables: {
          connections: [
            ConnectionHandler.getConnectionID(
              data.node.id!,
              "VendorView_riskAssessments",
            ),
          ],
          input: {
            vendorId: data.node.id!,
            assessedBy: data.viewer.user.people?.id!,
            expiresAt: nextYear.toISOString(),
            dataSensitivity: formValues.dataSensitivity,
            businessImpact: formValues.businessImpact,
            notes: formValues.notes || "Risk assessment",
          },
        },
        onCompleted: () => {
          toast({
            title: "Success",
            description: "Risk assessment created successfully",
            variant: "default",
          });
          setShowRiskAssessmentModal(false);
          loadQuery(
            {
              vendorId: data.node.id!,
              organizationId: organizationId!,
            },
            {
              fetchPolicy: "network-only",
            },
          );
        },
        onError: (error) => {
          toast({
            title: "Error",
            description: error.message || "Failed to create risk assessment",
            variant: "destructive",
          });
        },
      });
    },
    [
      createRiskAssessment,
      data.node.id,
      data.viewer.user.people?.id,
      loadQuery,
      toast,
      organizationId,
    ],
  );

  const formatDate = (dateStr: string | null | undefined) => {
    if (!dateStr) return "N/A";
    return new Date(dateStr).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  const renderTabContent = () => {
    switch (activeTab) {
      case "overview":
        return (
          <div className="space-y-6">
            <div className="flex flex-wrap gap-x-12 gap-y-4 items-start mb-6">
              <div className="space-y-1">
                <p className="text-sm font-medium text-secondary">Vendor ID</p>
                <p className="text-sm text-tertiary">{data.node.id}</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm font-medium text-secondary">Joined</p>
                <p className="text-sm text-tertiary">
                  {formatDate(data.node.createdAt)}
                </p>
              </div>
            </div>

            <div className="grid grid-cols-1 gap-6">
              {/* Vendor details card */}
              <div className="rounded-xl border border-border bg-white p-0 overflow-hidden">
                <div className="px-5 py-4 border-b border-[rgba(2,42,2,0.08)]">
                  <h3 className="text-base font-medium">Vendor details</h3>
                </div>
                <div className="p-5 space-y-4">
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-[#6B716A]">Name</p>
                    <div className="rounded-lg bg-[rgba(5,77,5,0.03)] px-2 py-1.5">
                      <input
                        type="text"
                        value={formData.name}
                        onChange={(e) =>
                          handleFieldChange("name", e.target.value)
                        }
                        className="w-full font-geist font-medium text-[16px] leading-[1.5em] text-[#141E12] bg-transparent border-0 outline-none focus:ring-0 focus-visible:ring-0 p-0"
                      />
                    </div>
                  </div>

                  <div className="space-y-1">
                    <p className="text-sm font-medium text-[#6B716A]">
                      Description
                    </p>
                    <div className="rounded-lg bg-[rgba(5,77,5,0.03)] px-2 py-1.5">
                      <textarea
                        value={formData.description}
                        onChange={(e) => {
                          handleFieldChange("description", e.target.value);
                          e.target.style.height = "0";
                          e.target.style.height = e.target.scrollHeight + "px";
                        }}
                        className="w-full font-geist font-medium text-[16px] leading-[1.5em] text-[#141E12] bg-transparent border-0 resize-none outline-none focus:ring-0 focus-visible:ring-0 p-0"
                        style={{
                          overflow: "hidden",
                          height: "auto",
                          minHeight: "1.5em",
                        }}
                      />
                    </div>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-[#6B716A]">Category</p>
                    <div className="rounded-lg bg-[rgba(5,77,5,0.03)] px-2 py-1.5">
                      <select
                        value={formData.category || ""}
                        onChange={(e) => {
                          const value = e.target.value;
                          handleFieldChange("category", value ? (value as VendorCategory) : null);
                        }}
                        className="w-full font-geist font-medium text-[16px] leading-[1.5em] text-[#141E12] bg-transparent border-0 outline-none focus:ring-0 focus-visible:ring-0 p-0"
                      >
                        {VENDOR_CATEGORIES.map(({ value, label }) => (
                          <option key={value} value={value}>
                            {label}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-[#6B716A]">
                      Legal name
                    </p>
                    <div className="rounded-lg bg-[rgba(5,77,5,0.03)] px-2 py-1.5">
                      <input
                        type="text"
                        value={formData.legalName}
                        onChange={(e) =>
                          handleFieldChange("legalName", e.target.value)
                        }
                        className="w-full font-geist font-medium text-[16px] leading-[1.5em] text-[#141E12] bg-transparent border-0 outline-none focus:ring-0 focus-visible:ring-0 p-0"
                      />
                    </div>
                  </div>

                  <div className="space-y-1">
                    <p className="text-sm font-medium text-[#6B716A]">
                      Headquarter Address
                    </p>
                    <div className="rounded-lg bg-[rgba(5,77,5,0.03)] px-2 py-1.5">
                      <input
                        type="text"
                        value={formData.headquarterAddress}
                        onChange={(e) =>
                          handleFieldChange(
                            "headquarterAddress",
                            e.target.value,
                          )
                        }
                        className="w-full font-geist font-medium text-[16px] leading-[1.5em] text-[#141E12] bg-transparent border-0 outline-none focus:ring-0 focus-visible:ring-0 p-0"
                      />
                    </div>
                  </div>

                  <div className="space-y-1">
                    <p className="text-sm font-medium text-[#6B716A]">
                      Website URL
                    </p>
                    <div className="rounded-lg bg-[rgba(5,77,5,0.03)] px-2 py-1.5">
                      <input
                        type="text"
                        value={formData.websiteUrl}
                        onChange={(e) =>
                          handleFieldChange("websiteUrl", e.target.value)
                        }
                        className="w-full font-geist font-medium text-[16px] leading-[1.5em] text-[#141E12] bg-transparent border-0 outline-none focus:ring-0 focus-visible:ring-0 p-0"
                      />
                    </div>
                  </div>
                </div>
              </div>

              {/* Ownership details card */}
              <div className="rounded-xl border border-border bg-white p-0 overflow-hidden">
                <div className="px-5 py-4 border-b border-[rgba(2,42,2,0.08)]">
                  <h3 className="text-base font-medium">Ownership details</h3>
                </div>
                <div className="p-5 space-y-4">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium text-[#6B716A]">
                        Business Owner
                      </p>
                      <HelpCircle className="h-4 w-4 text-tertiary" />
                    </div>
                    <div className="bg-[rgba(5,77,5,0.03)] rounded-lg p-1">
                      <PeopleSelector
                        organizationRef={data.organization}
                        selectedPersonId={formData.businessOwnerId}
                        onSelect={(value) =>
                          handleFieldChange("businessOwnerId", value)
                        }
                        placeholder="Select"
                      />
                    </div>
                  </div>

                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium text-[#6B716A]">
                        Security Owner
                      </p>
                      <HelpCircle className="h-4 w-4 text-tertiary" />
                    </div>
                    <div className="bg-[rgba(5,77,5,0.03)] rounded-lg p-1">
                      <PeopleSelector
                        organizationRef={data.organization}
                        selectedPersonId={formData.securityOwnerId}
                        onSelect={(value) =>
                          handleFieldChange("securityOwnerId", value)
                        }
                        placeholder="Select"
                      />
                    </div>
                  </div>
                </div>
              </div>

              {/* Links card */}
              <div className="rounded-xl border border-border bg-white p-0 overflow-hidden">
                <div className="px-5 py-4 border-b border-[rgba(2,42,2,0.08)]">
                  <h3 className="text-base font-medium">Links</h3>
                </div>
                <div className="divide-y">
                  <div className="grid grid-cols-1 md:grid-cols-[180px_1fr] py-4 px-5">
                    <p className="text-sm font-medium text-[#6B716A]">
                      Status Page URL
                    </p>
                    <div className="flex items-center space-x-2 w-full">
                      <Input
                        value={formData.statusPageUrl}
                        onChange={(e) =>
                          handleFieldChange("statusPageUrl", e.target.value)
                        }
                        className="border-0 bg-transparent p-0 shadow-none focus-visible:ring-0 flex-1"
                      />
                      {formData.statusPageUrl && (
                        <a
                          href={formData.statusPageUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                        >
                          <svg
                            width="16"
                            height="16"
                            viewBox="0 0 16 16"
                            fill="none"
                            xmlns="http://www.w3.org/2000/svg"
                          >
                            <path
                              d="M6.66667 9.33333L13.3333 2.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M8.66667 2.66667H13.3333V7.33333"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M13.3333 9.33333V12.6667C13.3333 13.0203 13.1929 13.3594 12.9428 13.6095C12.6928 13.8595 12.3536 14 12 14H3.33333C2.97971 14 2.64057 13.8595 2.39052 13.6095C2.14048 13.3594 2 13.0203 2 12.6667V4C2 3.64638 2.14048 3.30724 2.39052 3.05719C2.64057 2.80714 2.97971 2.66667 3.33333 2.66667H6.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                          </svg>
                        </a>
                      )}
                    </div>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-[180px_1fr] py-4 px-5">
                    <p className="text-sm font-medium text-[#6B716A]">
                      Terms of Service URL
                    </p>
                    <div className="flex items-center space-x-2 w-full">
                      <Input
                        value={formData.termsOfServiceUrl}
                        onChange={(e) =>
                          handleFieldChange("termsOfServiceUrl", e.target.value)
                        }
                        className="border-0 bg-transparent p-0 shadow-none focus-visible:ring-0 flex-1"
                      />
                      {formData.termsOfServiceUrl && (
                        <a
                          href={formData.termsOfServiceUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                        >
                          <svg
                            width="16"
                            height="16"
                            viewBox="0 0 16 16"
                            fill="none"
                            xmlns="http://www.w3.org/2000/svg"
                          >
                            <path
                              d="M6.66667 9.33333L13.3333 2.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M8.66667 2.66667H13.3333V7.33333"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M13.3333 9.33333V12.6667C13.3333 13.0203 13.1929 13.3594 12.9428 13.6095C12.6928 13.8595 12.3536 14 12 14H3.33333C2.97971 14 2.64057 13.8595 2.39052 13.6095C2.14048 13.3594 2 13.0203 2 12.6667V4C2 3.64638 2.14048 3.30724 2.39052 3.05719C2.64057 2.80714 2.97971 2.66667 3.33333 2.66667H6.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                          </svg>
                        </a>
                      )}
                    </div>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-[180px_1fr] py-4 px-5">
                    <p className="text-sm font-medium text-[#6B716A]">
                      Privacy Policy URL
                    </p>
                    <div className="flex items-center space-x-2 w-full">
                      <Input
                        value={formData.privacyPolicyUrl}
                        onChange={(e) =>
                          handleFieldChange("privacyPolicyUrl", e.target.value)
                        }
                        className="border-0 bg-transparent p-0 shadow-none focus-visible:ring-0 flex-1"
                      />
                      {formData.privacyPolicyUrl && (
                        <a
                          href={formData.privacyPolicyUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                        >
                          <svg
                            width="16"
                            height="16"
                            viewBox="0 0 16 16"
                            fill="none"
                            xmlns="http://www.w3.org/2000/svg"
                          >
                            <path
                              d="M6.66667 9.33333L13.3333 2.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M8.66667 2.66667H13.3333V7.33333"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M13.3333 9.33333V12.6667C13.3333 13.0203 13.1929 13.3594 12.9428 13.6095C12.6928 13.8595 12.3536 14 12 14H3.33333C2.97971 14 2.64057 13.8595 2.39052 13.6095C2.14048 13.3594 2 13.0203 2 12.6667V4C2 3.64638 2.14048 3.30724 2.39052 3.05719C2.64057 2.80714 2.97971 2.66667 3.33333 2.66667H6.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                          </svg>
                        </a>
                      )}
                    </div>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-[180px_1fr] py-4 px-5">
                    <p className="text-sm font-medium text-[#6B716A]">
                      Service Level Agreement URL
                    </p>
                    <div className="flex items-center space-x-2 w-full">
                      <Input
                        value={formData.serviceLevelAgreementUrl}
                        onChange={(e) =>
                          handleFieldChange(
                            "serviceLevelAgreementUrl",
                            e.target.value,
                          )
                        }
                        className="border-0 bg-transparent p-0 shadow-none focus-visible:ring-0 flex-1"
                      />
                      {formData.serviceLevelAgreementUrl && (
                        <a
                          href={formData.serviceLevelAgreementUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                        >
                          <svg
                            width="16"
                            height="16"
                            viewBox="0 0 16 16"
                            fill="none"
                            xmlns="http://www.w3.org/2000/svg"
                          >
                            <path
                              d="M6.66667 9.33333L13.3333 2.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M8.66667 2.66667H13.3333V7.33333"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M13.3333 9.33333V12.6667C13.3333 13.0203 13.1929 13.3594 12.9428 13.6095C12.6928 13.8595 12.3536 14 12 14H3.33333C2.97971 14 2.64057 13.8595 2.39052 13.6095C2.14048 13.3594 2 13.0203 2 12.6667V4C2 3.64638 2.14048 3.30724 2.39052 3.05719C2.64057 2.80714 2.97971 2.66667 3.33333 2.66667H6.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                          </svg>
                        </a>
                      )}
                    </div>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-[180px_1fr] py-4 px-5">
                    <p className="text-sm font-medium text-[#6B716A]">
                      Business Associate Agreement URL
                    </p>
                    <div className="flex items-center space-x-2 w-full">
                      <Input
                        value={formData.businessAssociateAgreementUrl}
                        onChange={(e) =>
                          handleFieldChange(
                            "businessAssociateAgreementUrl",
                            e.target.value,
                          )
                        }
                        className="border-0 bg-transparent p-0 shadow-none focus-visible:ring-0 flex-1"
                      />
                      {formData.businessAssociateAgreementUrl && (
                        <a
                          href={formData.businessAssociateAgreementUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                        >
                          <svg
                            width="16"
                            height="16"
                            viewBox="0 0 16 16"
                            fill="none"
                            xmlns="http://www.w3.org/2000/svg"
                          >
                            <path
                              d="M6.66667 9.33333L13.3333 2.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M8.66667 2.66667H13.3333V7.33333"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M13.3333 9.33333V12.6667C13.3333 13.0203 13.1929 13.3594 12.9428 13.6095C12.6928 13.8595 12.3536 14 12 14H3.33333C2.97971 14 2.64057 13.8595 2.39052 13.6095C2.14048 13.3594 2 13.0203 2 12.6667V4C2 3.64638 2.14048 3.30724 2.39052 3.05719C2.64057 2.80714 2.97971 2.66667 3.33333 2.66667H6.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                          </svg>
                        </a>
                      )}
                    </div>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-[180px_1fr] py-4 px-5">
                    <p className="text-sm font-medium text-[#6B716A]">
                      Subprocessors List URL
                    </p>
                    <div className="flex items-center space-x-2 w-full">
                      <Input
                        value={formData.subprocessorsListUrl}
                        onChange={(e) =>
                          handleFieldChange(
                            "subprocessorsListUrl",
                            e.target.value,
                          )
                        }
                        className="border-0 bg-transparent p-0 shadow-none focus-visible:ring-0 flex-1"
                      />
                      {formData.subprocessorsListUrl && (
                        <a
                          href={formData.subprocessorsListUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                        >
                          <svg
                            width="16"
                            height="16"
                            viewBox="0 0 16 16"
                            fill="none"
                            xmlns="http://www.w3.org/2000/svg"
                          >
                            <path
                              d="M6.66667 9.33333L13.3333 2.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M8.66667 2.66667H13.3333V7.33333"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M13.3333 9.33333V12.6667C13.3333 13.0203 13.1929 13.3594 12.9428 13.6095C12.6928 13.8595 12.3536 14 12 14H3.33333C2.97971 14 2.64057 13.8595 2.39052 13.6095C2.14048 13.3594 2 13.0203 2 12.6667V4C2 3.64638 2.14048 3.30724 2.39052 3.05719C2.64057 2.80714 2.97971 2.66667 3.33333 2.66667H6.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                          </svg>
                        </a>
                      )}
                    </div>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-[180px_1fr] py-4 px-5">
                    <p className="text-sm font-medium text-[#6B716A]">
                      Data Processing Agreement URL
                    </p>
                    <div className="flex items-center space-x-2 w-full">
                      <Input
                        value={formData.dataProcessingAgreementUrl}
                        onChange={(e) =>
                          handleFieldChange(
                            "dataProcessingAgreementUrl",
                            e.target.value,
                          )
                        }
                        className="border-0 bg-transparent p-0 shadow-none focus-visible:ring-0 flex-1"
                      />
                      {formData.dataProcessingAgreementUrl && (
                        <a
                          href={formData.dataProcessingAgreementUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                        >
                          <svg
                            width="16"
                            height="16"
                            viewBox="0 0 16 16"
                            fill="none"
                            xmlns="http://www.w3.org/2000/svg"
                          >
                            <path
                              d="M6.66667 9.33333L13.3333 2.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M8.66667 2.66667H13.3333V7.33333"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M13.3333 9.33333V12.6667C13.3333 13.0203 13.1929 13.3594 12.9428 13.6095C12.6928 13.8595 12.3536 14 12 14H3.33333C2.97971 14 2.64057 13.8595 2.39052 13.6095C2.14048 13.3594 2 13.0203 2 12.6667V4C2 3.64638 2.14048 3.30724 2.39052 3.05719C2.64057 2.80714 2.97971 2.66667 3.33333 2.66667H6.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                          </svg>
                        </a>
                      )}
                    </div>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-[180px_1fr] py-4 px-5">
                    <p className="text-sm font-medium text-[#6B716A]">
                      Security Page URL
                    </p>
                    <div className="flex items-center space-x-2 w-full">
                      <Input
                        value={formData.securityPageUrl}
                        onChange={(e) =>
                          handleFieldChange("securityPageUrl", e.target.value)
                        }
                        className="border-0 bg-transparent p-0 shadow-none focus-visible:ring-0 flex-1"
                      />
                      {formData.securityPageUrl && (
                        <a
                          href={formData.securityPageUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                        >
                          <svg
                            width="16"
                            height="16"
                            viewBox="0 0 16 16"
                            fill="none"
                            xmlns="http://www.w3.org/2000/svg"
                          >
                            <path
                              d="M6.66667 9.33333L13.3333 2.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M8.66667 2.66667H13.3333V7.33333"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M13.3333 9.33333V12.6667C13.3333 13.0203 13.1929 13.3594 12.9428 13.6095C12.6928 13.8595 12.3536 14 12 14H3.33333C2.97971 14 2.64057 13.8595 2.39052 13.6095C2.14048 13.3594 2 13.0203 2 12.6667V4C2 3.64638 2.14048 3.30724 2.39052 3.05719C2.64057 2.80714 2.97971 2.66667 3.33333 2.66667H6.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                          </svg>
                        </a>
                      )}
                    </div>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-[180px_1fr] py-4 px-5">
                    <p className="text-sm font-medium text-[#6B716A]">
                      Trust Page URL
                    </p>
                    <div className="flex items-center space-x-2 w-full">
                      <Input
                        value={formData.trustPageUrl}
                        onChange={(e) =>
                          handleFieldChange("trustPageUrl", e.target.value)
                        }
                        className="border-0 bg-transparent p-0 shadow-none focus-visible:ring-0 flex-1"
                      />
                      {formData.trustPageUrl && (
                        <a
                          href={formData.trustPageUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                        >
                          <svg
                            width="16"
                            height="16"
                            viewBox="0 0 16 16"
                            fill="none"
                            xmlns="http://www.w3.org/2000/svg"
                          >
                            <path
                              d="M6.66667 9.33333L13.3333 2.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M8.66667 2.66667H13.3333V7.33333"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                            <path
                              d="M13.3333 9.33333V12.6667C13.3333 13.0203 13.1929 13.3594 12.9428 13.6095C12.6928 13.8595 12.3536 14 12 14H3.33333C2.97971 14 2.64057 13.8595 2.39052 13.6095C2.14048 13.3594 2 13.0203 2 12.6667V4C2 3.64638 2.14048 3.30724 2.39052 3.05719C2.64057 2.80714 2.97971 2.66667 3.33333 2.66667H6.66667"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                          </svg>
                        </a>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        );
      case "certifications":
        return (
          <CertificationsView
            certifications={[...formData.certifications]}
            onAdd={(cert) =>
              handleCertificationChange([
                ...formData.certifications.filter((c) => c !== cert),
                cert,
              ])
            }
            onRemove={(cert) =>
              handleCertificationChange(
                formData.certifications.filter((c) => c !== cert),
              )
            }
          />
        );
      case "complianceReports":
        return (
          <div className="space-y-4">
            <ComplianceReportsTable
              reports={
                data.node.complianceReports?.edges.map((edge) => edge.node) ??
                []
              }
              onDelete={handleDeleteReport}
              onUpload={handleUploadReport}
            />
          </div>
        );
      case "riskAssessments":
        return (
          <div className="space-y-4">
            <RiskAssessmentsTable
              assessments={data.node.riskAssessments?.edges || []}
              onCreateAssessment={handleCreateRiskAssessment}
            />
          </div>
        );
      default:
        return null;
    }
  };

  const handleAssessVendor = useCallback(() => {
    if (!websiteUrl) {
      toast({
        title: "Error",
        description: "Please enter a website URL",
        variant: "destructive",
      });
      return;
    }

    commitAssessVendor({
      variables: {
        input: {
          id: data.node.id!,
          websiteUrl,
        },
      },
      onCompleted: (response) => {
        const newData = response.assessVendor.vendor;
        const changedFields = new Set<string>();

        // Compare and track changes
        Object.entries(newData).forEach(([key, value]) => {
          if (value != null && key !== 'id' && formData[key as keyof typeof formData] !== value) {
            changedFields.add(key);
          }
        });

        setFormData(prevData => {
          const newFormData = {
            ...prevData,
            name: newData.name || prevData.name,
            description: newData.description || prevData.description,
            statusPageUrl: newData.statusPageUrl || prevData.statusPageUrl,
            termsOfServiceUrl: newData.termsOfServiceUrl || prevData.termsOfServiceUrl,
            privacyPolicyUrl: newData.privacyPolicyUrl || prevData.privacyPolicyUrl,
            serviceLevelAgreementUrl: newData.serviceLevelAgreementUrl || prevData.serviceLevelAgreementUrl,
            dataProcessingAgreementUrl: newData.dataProcessingAgreementUrl || prevData.dataProcessingAgreementUrl,
            businessAssociateAgreementUrl: newData.businessAssociateAgreementUrl || prevData.businessAssociateAgreementUrl,
            subprocessorsListUrl: newData.subprocessorsListUrl || prevData.subprocessorsListUrl,
            securityPageUrl: newData.securityPageUrl || prevData.securityPageUrl,
            trustPageUrl: newData.trustPageUrl || prevData.trustPageUrl,
            certifications: newData.certifications || prevData.certifications,
            headquarterAddress: newData.headquarterAddress || prevData.headquarterAddress,
            legalName: newData.legalName || prevData.legalName,
            websiteUrl: newData.websiteUrl || prevData.websiteUrl,
            category: newData.category || prevData.category,
          };
          return newFormData;
        });

        setEditedFields(prev => {
          const newSet = new Set(prev);
          changedFields.forEach(field => newSet.add(field));
          return newSet;
        });

        toast({
          title: "Success",
          description: changedFields.size > 0
            ? "Vendor has been assessed. Review and save the changes."
            : "Vendor data has been assessed. No changes were found.",
        });

        setIsAssessDialogOpen(false);
        setWebsiteUrl("");
      },
      onError: (error) => {
        toast({
          title: "Error",
          description: error.message,
          variant: "destructive",
        });
      },
    });
  }, [commitAssessVendor, data.node.id, websiteUrl, toast, formData]);
  return (
    <PageTemplate
      title={formData.name}
      actions={
        <Button
          className="rounded-full bg-[rgba(0,39,0,0.05)] text-[#141E12] hover:bg-[rgba(0,39,0,0.08)] h-8 px-3"
          size="sm"
          onClick={() => setIsAssessDialogOpen(true)}
        >
          Assessement From Website
        </Button>
      }
    >
      <div className="border-b mb-6">
        <div className="flex">
          <button
            onClick={() => setActiveTab("overview")}
            className={cn(
              "px-4 py-3 text-sm font-medium border-b-2 -mb-px",
              activeTab === "overview"
                ? "border-primary text-primary"
                : "border-transparent text-secondary hover:text-primary hover:border-secondary",
            )}
          >
            Overview
          </button>
          <button
            onClick={() => setActiveTab("certifications")}
            className={cn(
              "px-4 py-3 text-sm font-medium border-b-2 -mb-px",
              activeTab === "certifications"
                ? "border-primary text-primary"
                : "border-transparent text-secondary hover:text-primary hover:border-secondary",
            )}
          >
            Certifications
          </button>
          <button
            onClick={() => setActiveTab("complianceReports")}
            className={cn(
              "px-4 py-3 text-sm font-medium border-b-2 -mb-px",
              activeTab === "complianceReports"
                ? "border-primary text-primary"
                : "border-transparent text-secondary hover:text-primary hover:border-secondary",
            )}
          >
            Compliance reports
          </button>
          <button
            onClick={() => setActiveTab("riskAssessments")}
            className={cn(
              "px-4 py-3 text-sm font-medium border-b-2 -mb-px",
              activeTab === "riskAssessments"
                ? "border-primary text-primary"
                : "border-transparent text-secondary hover:text-primary hover:border-secondary",
            )}
          >
            Risk assessments
          </button>
        </div>
      </div>

      {/* Tab content */}
      {renderTabContent()}

      {/* Save/cancel buttons */}
      {hasChanges && (
        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" onClick={handleCancel}>
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            className="bg-primary text-invert hover:bg-primary/90"
          >
            Save Changes
          </Button>
        </div>
      )}

      {/* Risk Assessment Modal */}
      <RiskAssessmentModal
        isOpen={showRiskAssessmentModal}
        onClose={() => setShowRiskAssessmentModal(false)}
        onSubmit={handleRiskAssessmentSubmit}
        businessOwnerId={formData.businessOwnerId}
      />

      {/* Assessement From Website Modal */}
      <Dialog open={isAssessDialogOpen} onOpenChange={setIsAssessDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Assessement From Website</DialogTitle>
          </DialogHeader>
          <form onSubmit={(e) => {
            e.preventDefault();
            handleAssessVendor();
          }}>
            <div className="space-y-4">
              <Input
                type="url"
                placeholder="https://www.example.com"
                value={websiteUrl}
                onChange={(e) => setWebsiteUrl(e.target.value)}
                disabled={isInFlight}
                required
              />
              {isInFlight && (
                <div className="flex items-center justify-center text-sm text-[#6B716A]">
                  <svg
                    className="animate-spin -ml-1 mr-3 h-5 w-5 text-[#054D05]"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    ></circle>
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                  </svg>
                  Assessing vendor data...
                </div>
              )}
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <Button
                type="button"
                variant="outline"
                onClick={() => setIsAssessDialogOpen(false)}
                disabled={isInFlight}
              >
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={isInFlight}
                className={isInFlight ? "opacity-50 cursor-not-allowed" : ""}
              >
                {isInFlight ? "Processing..." : "Assess"}
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </PageTemplate>
  );
}

export default function VendorView() {
  const { vendorId, organizationId } = useParams();

  const [queryRef, loadQuery] =
    useQueryLoader<VendorViewQueryType>(vendorViewQuery);

  useEffect(() => {
    loadQuery({ vendorId: vendorId!, organizationId: organizationId! });
  }, [loadQuery, vendorId, organizationId]);

  if (!queryRef) {
    return <VendorViewSkeleton />;
  }

  return (
    <Suspense fallback={<VendorViewSkeleton />}>
      <VendorViewContent queryRef={queryRef} />
    </Suspense>
  );
}
