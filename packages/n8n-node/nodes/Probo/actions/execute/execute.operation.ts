import { parse, getOperationAST, type DocumentNode } from 'graphql';
import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Query',
		name: 'query',
		type: 'string',
		typeOptions: {
			rows: 5,
		},
		displayOptions: {
			show: {
				resource: ['execute'],
			},
		},
		default: '',
		description:
			'The complete GraphQL operation including operation name and variable declarations (e.g., "query GetUser($userId: ID!) { node(id: $userId) { id } }" or "mutation UpdateUser($input: UpdateUserInput!) { updateUser(input: $input) { id } }")',
		required: true,
	},
	{
		displayName: 'Variables',
		name: 'variables',
		type: 'json',
		displayOptions: {
			show: {
				resource: ['execute'],
			},
		},
		default: '{}',
		description: 'GraphQL variables as JSON object',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const query = this.getNodeParameter('query', itemIndex) as string;
	const variablesParam = this.getNodeParameter('variables', itemIndex) as string;

	let document: DocumentNode;
	try {
		document = parse(query);
	} catch (error) {
		throw new Error(
			`Invalid GraphQL operation: ${error instanceof Error ? error.message : String(error)}`,
		);
	}

	const operationAST = getOperationAST(document);
	if (!operationAST) {
		throw new Error('GraphQL operation must contain a query, mutation, or subscription');
	}

	if (!operationAST.name) {
		throw new Error(
			'GraphQL operation must have a name (e.g., "query GetUser { ... }" or "mutation UpdateUser { ... }")',
		);
	}

	let variables = {};
	if (variablesParam) {
		try {
			variables =
				typeof variablesParam === 'string' ? JSON.parse(variablesParam) : variablesParam;
		} catch (error) {
			throw new Error(`Invalid JSON in Variables: ${error}`);
		}
	}

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
