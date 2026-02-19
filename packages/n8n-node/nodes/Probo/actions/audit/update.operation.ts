import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Audit ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['audit'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the audit to update',
		required: true,
	},
	{
		displayName: 'Update Fields',
		name: 'updateFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['audit'],
				operation: ['update'],
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
	const id = this.getNodeParameter('id', itemIndex) as string;
	const updateFields = this.getNodeParameter('updateFields', itemIndex, {}) as {
		name?: string;
		frameworkType?: string;
		state?: string;
		validFrom?: string;
		validUntil?: string;
	};

	const query = `
		mutation UpdateAudit($input: UpdateAuditInput!) {
			updateAudit(input: $input) {
				audit {
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
	`;

	const input: Record<string, unknown> = { id };
	if (updateFields.name) input.name = updateFields.name;
	if (updateFields.frameworkType) input.frameworkType = updateFields.frameworkType;
	if (updateFields.state) input.state = updateFields.state;
	if (updateFields.validFrom) input.validFrom = updateFields.validFrom;
	if (updateFields.validUntil) input.validUntil = updateFields.validUntil;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
