import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "../useMutationWithToasts";
import type { SAMLConfigurationGraphCreateMutation } from "./__generated__/SAMLConfigurationGraphCreateMutation.graphql";
import type { SAMLConfigurationGraphUpdateMutation } from "./__generated__/SAMLConfigurationGraphUpdateMutation.graphql";
import type { SAMLConfigurationGraphDeleteMutation } from "./__generated__/SAMLConfigurationGraphDeleteMutation.graphql";
import type { SAMLConfigurationGraphEnableMutation } from "./__generated__/SAMLConfigurationGraphEnableMutation.graphql";
import type { SAMLConfigurationGraphDisableMutation } from "./__generated__/SAMLConfigurationGraphDisableMutation.graphql";
import type { SAMLConfigurationGraphInitiateDomainVerificationMutation } from "./__generated__/SAMLConfigurationGraphInitiateDomainVerificationMutation.graphql";
import type { SAMLConfigurationGraphVerifyDomainMutation } from "./__generated__/SAMLConfigurationGraphVerifyDomainMutation.graphql";

const createSAMLConfigurationMutation = graphql`
  mutation SAMLConfigurationGraphCreateMutation(
    $input: CreateSAMLConfigurationInput!
  ) {
    createSAMLConfiguration(input: $input) {
      samlConfiguration {
        id
        enabled
        emailDomain
        enforcementPolicy
        spEntityId
        spAcsUrl
        spMetadataUrl
        testLoginUrl
        idpEntityId
        idpSsoUrl
        idpCertificate
        idpMetadataUrl
        attributeEmail
        attributeFirstname
        attributeLastname
        attributeRole
        defaultRole
        autoSignupEnabled
        createdAt
        updatedAt
      }
    }
  }
`;

const updateSAMLConfigurationMutation = graphql`
  mutation SAMLConfigurationGraphUpdateMutation(
    $input: UpdateSAMLConfigurationInput!
  ) {
    updateSAMLConfiguration(input: $input) {
      samlConfiguration {
        id
        enabled
        emailDomain
        enforcementPolicy
        spEntityId
        spAcsUrl
        spMetadataUrl
        testLoginUrl
        idpEntityId
        idpSsoUrl
        idpCertificate
        idpMetadataUrl
        attributeEmail
        attributeFirstname
        attributeLastname
        attributeRole
        defaultRole
        autoSignupEnabled
        createdAt
        updatedAt
      }
    }
  }
`;

const deleteSAMLConfigurationMutation = graphql`
  mutation SAMLConfigurationGraphDeleteMutation(
    $input: DeleteSAMLConfigurationInput!
  ) {
    deleteSAMLConfiguration(input: $input) {
      deletedSAMLConfigurationId
    }
  }
`;

const enableSAMLMutation = graphql`
  mutation SAMLConfigurationGraphEnableMutation($input: EnableSAMLInput!) {
    enableSAML(input: $input) {
      samlConfiguration {
        id
        enabled
      }
    }
  }
`;

const disableSAMLMutation = graphql`
  mutation SAMLConfigurationGraphDisableMutation($input: DisableSAMLInput!) {
    disableSAML(input: $input) {
      samlConfiguration {
        id
        enabled
      }
    }
  }
`;

const initiateDomainVerificationMutation = graphql`
  mutation SAMLConfigurationGraphInitiateDomainVerificationMutation(
    $input: InitiateDomainVerificationInput!
  ) {
    initiateDomainVerification(input: $input) {
      samlConfiguration {
        id
        emailDomain
        domainVerified
        domainVerificationToken
        domainVerifiedAt
      }
      dnsRecord
    }
  }
`;

const verifyDomainMutation = graphql`
  mutation SAMLConfigurationGraphVerifyDomainMutation(
    $input: VerifyDomainInput!
  ) {
    verifyDomain(input: $input) {
      samlConfiguration {
        id
        domainVerified
        domainVerifiedAt
      }
      verified
    }
  }
`;

export function useCreateSAMLConfigurationMutation() {
  return useMutationWithToasts<SAMLConfigurationGraphCreateMutation>(
    createSAMLConfigurationMutation,
    {
      successMessage: "SAML configuration created successfully.",
      errorMessage: "Failed to create SAML configuration. Please try again.",
    }
  );
}

export function useUpdateSAMLConfigurationMutation() {
  return useMutationWithToasts<SAMLConfigurationGraphUpdateMutation>(
    updateSAMLConfigurationMutation,
    {
      successMessage: "SAML configuration updated successfully.",
      errorMessage: "Failed to update SAML configuration. Please try again.",
    }
  );
}

export function useDeleteSAMLConfigurationMutation() {
  return useMutationWithToasts<SAMLConfigurationGraphDeleteMutation>(
    deleteSAMLConfigurationMutation,
    {
      successMessage: "SAML configuration deleted successfully.",
      errorMessage: "Failed to delete SAML configuration. Please try again.",
    }
  );
}

export function useEnableSAMLMutation() {
  return useMutationWithToasts<SAMLConfigurationGraphEnableMutation>(
    enableSAMLMutation,
    {
      successMessage: "SAML enabled successfully.",
      errorMessage: "Failed to enable SAML. Please try again.",
    }
  );
}

export function useDisableSAMLMutation() {
  return useMutationWithToasts<SAMLConfigurationGraphDisableMutation>(
    disableSAMLMutation,
    {
      successMessage: "SAML disabled successfully.",
      errorMessage: "Failed to disable SAML. Please try again.",
    }
  );
}

export function useInitiateDomainVerificationMutation() {
  return useMutationWithToasts<SAMLConfigurationGraphInitiateDomainVerificationMutation>(
    initiateDomainVerificationMutation,
    {
      successMessage: "Domain verification initiated. Please add the DNS record.",
      errorMessage: "Failed to initiate domain verification. Please try again.",
    }
  );
}

export function useVerifyDomainMutation() {
  return useMutationWithToasts<SAMLConfigurationGraphVerifyDomainMutation>(
    verifyDomainMutation,
    {
      successMessage: "Domain verified successfully!",
      errorMessage: "Domain verification failed. Please ensure the DNS record is properly configured.",
    }
  );
}
