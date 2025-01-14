# DRP Template

### **Overview and Objectives**

- **Purpose**: Provide step-by-step instructions to recover **[Critical
  Services]** in case of a disaster, ensuring minimum downtime and data
  protection.
- **Scope**: This plan applies to **[Specify Services, e.g., Cloud SQL,
  application servers, etc.]**.
- **Objectives**: Restore critical services within **[Define Recovery Time
  Objective, e.g., 2 hours]**.

### Instructions

<aside>
⚠️

- Don’t overthink it.
- Unless specifically required by a customer, don’t change your cloud provider
for the recovery
</aside>

1. Create a VM
2. Run the following commands to log-in and to install the minimum software

   `your commands`

3. Allow the VM to connect to the Database

   `your commands`

4. Check the service is running

### **Troubleshooting**

> In this section provide everything you think it can help people running the
> instruction e.g. “Connection Refused”, “Docker Image does not exist”, etc.

### **Contact Information**

| Role                   | Name        | Contact Info           | Backup Contact        |
| ---------------------- | ----------- | ---------------------- | --------------------- |
| CTO                    | [Name]      | [Email, Phone]         | [Backup Contact]      |
| Cloud Provider Support | GCP Support | [Support Contact Info] | [Alternative Contact] |

### **Testing and Review Log**

| Date         | Test Type               | Result     | Notes/Improvements                    |
| ------------ | ----------------------- | ---------- | ------------------------------------- |
| [YYYY-MM-DD] | Backup Restoration Test | Successful | Updated recovery steps for [Service]. |
