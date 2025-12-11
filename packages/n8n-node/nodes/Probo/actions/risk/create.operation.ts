import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the risk',
		required: true,
	},
	{
		displayName: 'Category',
		name: 'category',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The category of the risk',
		required: true,
	},
	{
		displayName: 'Treatment',
		name: 'treatment',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['create'],
			},
		},
		options: [
			{
				name: 'Mitigated',
				value: 'MITIGATED',
			},
			{
				name: 'Accepted',
				value: 'ACCEPTED',
			},
			{
				name: 'Avoided',
				value: 'AVOIDED',
			},
			{
				name: 'Transferred',
				value: 'TRANSFERRED',
			},
		],
		default: 'MITIGATED',
		description: 'The treatment strategy for the risk',
		required: true,
	},
	{
		displayName: 'Inherent Likelihood',
		name: 'inherentLikelihood',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['create'],
			},
		},
		typeOptions: {
			minValue: 1,
			maxValue: 5,
		},
		default: 1,
		description: 'The inherent likelihood of the risk (1-5)',
		required: true,
	},
	{
		displayName: 'Inherent Impact',
		name: 'inherentImpact',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['create'],
			},
		},
		typeOptions: {
			minValue: 1,
			maxValue: 5,
		},
		default: 1,
		description: 'The inherent impact of the risk (1-5)',
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
				resource: ['risk'],
				operation: ['create'],
			},
		},
		options: [
			{
				displayName: 'Description',
				name: 'description',
				type: 'string',
				default: '',
				description: 'The description of the risk',
			},
			{
				displayName: 'Note',
				name: 'note',
				type: 'string',
				default: '',
				description: 'Additional notes about the risk',
			},
			{
				displayName: 'Owner ID',
				name: 'ownerId',
				type: 'string',
				default: '',
				description: 'The ID of the person who owns this risk',
			},
			{
				displayName: 'Residual Impact',
				name: 'residualImpact',
				type: 'number',
				typeOptions: {
					minValue: 1,
					maxValue: 5,
				},
				default: 1,
				description: 'The residual impact of the risk (1-5)',
			},
			{
				displayName: 'Residual Likelihood',
				name: 'residualLikelihood',
				type: 'number',
				typeOptions: {
					minValue: 1,
					maxValue: 5,
				},
				default: 1,
				description: 'The residual likelihood of the risk (1-5)',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const category = this.getNodeParameter('category', itemIndex) as string;
	const treatment = this.getNodeParameter('treatment', itemIndex) as string;
	const inherentLikelihood = this.getNodeParameter('inherentLikelihood', itemIndex) as number;
	const inherentImpact = this.getNodeParameter('inherentImpact', itemIndex) as number;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		description?: string;
		ownerId?: string;
		residualLikelihood?: number;
		residualImpact?: number;
		note?: string;
	};

	const query = `
		mutation CreateRisk($input: CreateRiskInput!) {
			createRisk(input: $input) {
				riskEdge {
					node {
						id
						name
						description
						category
						treatment
						inherentLikelihood
						inherentImpact
						inherentRiskScore
						residualLikelihood
						residualImpact
						residualRiskScore
						note
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		organizationId,
		name,
		category,
		treatment,
		inherentLikelihood,
		inherentImpact,
	};
	if (additionalFields.description) input.description = additionalFields.description;
	if (additionalFields.ownerId) input.ownerId = additionalFields.ownerId;
	if (additionalFields.residualLikelihood !== undefined) input.residualLikelihood = additionalFields.residualLikelihood;
	if (additionalFields.residualImpact !== undefined) input.residualImpact = additionalFields.residualImpact;
	if (additionalFields.note) input.note = additionalFields.note;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
