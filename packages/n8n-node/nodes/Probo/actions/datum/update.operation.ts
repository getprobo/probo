import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Datum ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the datum to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the datum',
	},
	{
		displayName: 'Data Classification',
		name: 'dataClassification',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: 'Public',
				value: 'PUBLIC',
			},
			{
				name: 'Internal',
				value: 'INTERNAL',
			},
			{
				name: 'Confidential',
				value: 'CONFIDENTIAL',
			},
			{
				name: 'Secret',
				value: 'SECRET',
			},
		],
		default: '',
		description: 'The classification of the data',
	},
	{
		displayName: 'Owner ID',
		name: 'ownerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the owner (People)',
	},
	{
		displayName: 'Vendor IDs',
		name: 'vendorIds',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'Comma-separated list of vendor IDs',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const dataClassification = this.getNodeParameter('dataClassification', itemIndex, '') as string;
	const ownerId = this.getNodeParameter('ownerId', itemIndex, '') as string;
	const vendorIdsStr = this.getNodeParameter('vendorIds', itemIndex, '') as string;

	const query = `
		mutation UpdateDatum($input: UpdateDatumInput!) {
			updateDatum(input: $input) {
				datum {
					id
					name
					dataClassification
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, any> = { id };
	if (name) input.name = name;
	if (dataClassification) input.dataClassification = dataClassification;
	if (ownerId) input.ownerId = ownerId;
	if (vendorIdsStr) {
		const vendorIds = vendorIdsStr.split(',').map((vid) => vid.trim()).filter(Boolean);
		if (vendorIds.length > 0) input.vendorIds = vendorIds;
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

