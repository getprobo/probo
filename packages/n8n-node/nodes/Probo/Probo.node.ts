import {
	NodeConnectionTypes,
	type IExecuteFunctions,
	type INodeExecutionData,
	type INodeType,
	type INodeTypeDescription,
} from 'n8n-workflow';
import {
	getAllResourceOperations,
	getAllResourceFields,
	getExecuteFunction,
} from './actions';

export class Probo implements INodeType {
	description: INodeTypeDescription = {
		displayName: 'Probo',
		name: 'probo',
		icon: { light: 'file:../../icons/probo-light.svg', dark: 'file:../../icons/probo.svg' },
		group: ['input'],
		version: 1,
		subtitle: '={{$parameter["resource"]}} / {{$parameter["operation"]}}',
		description: 'Consume data from the Probo API',
		defaults: {
			name: 'Probo',
		},
		usableAsTool: true,
		inputs: [NodeConnectionTypes.Main],
		outputs: [NodeConnectionTypes.Main],
		credentials: [
			{
				name: 'proboApi',
				required: true,
				displayOptions: {
					show: {
						authentication: ['apiKey'],
					},
				},
			},
		],
		requestDefaults: {
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json',
			},
		},
		properties: [
			{
				displayName: 'Authentication',
				name: 'authentication',
				type: 'options',
				options: [
					{
						name: 'API Key',
						value: 'apiKey',
					},
				],
				default: 'apiKey',
			},
			{
				displayName: 'Resource',
				name: 'resource',
				type: 'options',
				noDataExpression: true,
				options: [
					{
						name: 'Asset',
						value: 'asset',
						description: 'Manage assets',
					},
					{
						name: 'Control',
						value: 'control',
						description: 'Manage controls',
					},
					{
						name: 'Data',
						value: 'datum',
						description: 'Manage data',
					},
					{
						name: 'Execute',
						value: 'execute',
						description: 'Execute a GraphQL query or mutation',
					},
					{
						name: 'Framework',
						value: 'framework',
						description: 'Manage frameworks',
					},
					{
						name: 'Measure',
						value: 'measure',
						description: 'Manage measures',
					},
					{
						name: 'Meeting',
						value: 'meeting',
						description: 'Manage meetings',
					},
					{
						name: 'Member',
						value: 'member',
						description: 'Manage organization members',
					},
					{
						name: 'Organization',
						value: 'organization',
						description: 'Manage organizations',
					},
					{
						name: 'People',
						value: 'people',
						description: 'Manage people',
					},
					{
						name: 'Risk',
						value: 'risk',
						description: 'Manage risks',
					},
					{
						name: 'Vendor',
						value: 'vendor',
						description: 'Manage vendors',
					},
				],
				default: 'execute',
			},
			...getAllResourceOperations(),
			...getAllResourceFields(),
		],
	};

	async execute(this: IExecuteFunctions): Promise<INodeExecutionData[][]> {
		const items = this.getInputData();
		const returnData: INodeExecutionData[] = [];

		for (let i = 0; i < items.length; i++) {
			const resource = this.getNodeParameter('resource', i) as string;
			const operation = this.getNodeParameter('operation', i, 'execute') as string;

			const executeFunction = getExecuteFunction(resource, operation);
			const result = await executeFunction.call(this, i);
			returnData.push(result);
		}

		return [returnData];
	}

	methods = {
		listSearch: {},
	};
}
