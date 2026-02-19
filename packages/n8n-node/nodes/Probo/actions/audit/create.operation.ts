import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['audit'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Framework ID',
		name: 'frameworkId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['audit'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the framework',
		required: true,
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['audit'],
				operation: ['create'],
			},
		},
		options: [
			{
				displayName: 'Framework Type',
				name: 'frameworkType',
				type: 'string',
				default: '',
				description: 'The framework type of the audit',
			},
			{
				displayName: 'Name',
				name: 'name',
				type: 'string',
				default: '',
				description: 'The name of the audit',
			},
			{
				displayName: 'State',
				name: 'state',
				type: 'options',
				options: [
					{
						name: 'Completed',
						value: 'COMPLETED',
					},
					{
						name: 'In Progress',
						value: 'IN_PROGRESS',
					},
					{
						name: 'Not Started',
						value: 'NOT_STARTED',
					},
					{
						name: 'Outdated',
						value: 'OUTDATED',
					},
					{
						name: 'Rejected',
						value: 'REJECTED',
					},
				],
				default: 'NOT_STARTED',
				description: 'The state of the audit',
			},
			{
				displayName: 'Valid From',
				name: 'validFrom',
				type: 'dateTime',
				default: '',
				description: 'The start date of the audit validity period',
			},
			{
				displayName: 'Valid Until',
				name: 'validUntil',
				type: 'dateTime',
				default: '',
				description: 'The end date of the audit validity period',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const frameworkId = this.getNodeParameter('frameworkId', itemIndex) as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		name?: string;
		frameworkType?: string;
		state?: string;
		validFrom?: string;
		validUntil?: string;
	};

	const query = `
		mutation CreateAudit($input: CreateAuditInput!) {
			createAudit(input: $input) {
				auditEdge {
					node {
						id
						name
						frameworkType
						state
						validFrom
						validUntil
						reportUrl
						trustCenterVisibility
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		organizationId,
		frameworkId,
	};
	if (additionalFields.name) input.name = additionalFields.name;
	if (additionalFields.frameworkType) input.frameworkType = additionalFields.frameworkType;
	if (additionalFields.state) input.state = additionalFields.state;
	if (additionalFields.validFrom) input.validFrom = additionalFields.validFrom;
	if (additionalFields.validUntil) input.validUntil = additionalFields.validUntil;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
