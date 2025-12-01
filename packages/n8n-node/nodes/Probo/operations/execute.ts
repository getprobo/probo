import { parse, getOperationAST, type DocumentNode } from 'graphql';
import type { IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';

export interface GraphQLParameters {
	query: string;
	variables?: string;
}

export async function executeGraphQL(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const query = this.getNodeParameter('query', itemIndex) as string;
	const variablesParam = this.getNodeParameter('variables', itemIndex) as string;

	let document: DocumentNode;
	try {
		document = parse(query);
	} catch (error) {
		throw new Error(`Invalid GraphQL operation: ${error instanceof Error ? error.message : String(error)}`);
	}

	const operationAST = getOperationAST(document);
	if (!operationAST) {
		throw new Error('GraphQL operation must contain a query, mutation, or subscription');
	}

	if (!operationAST.name) {
		throw new Error('GraphQL operation must have a name (e.g., "query GetUser { ... }" or "mutation UpdateUser { ... }")');
	}

	const operationName = operationAST.name.value;

	let variables = {};
	if (variablesParam) {
		try {
			variables =
				typeof variablesParam === 'string' ? JSON.parse(variablesParam) : variablesParam;
		} catch (error) {
			throw new Error(`Invalid JSON in Variables: ${error}`);
		}
	}

	const body: {
		query: string;
		operationName?: string;
		variables?: Record<string, any>;
	} = {
		query,
		operationName,
	};

	if (Object.keys(variables).length > 0) {
		body.variables = variables;
	}

	const credentials = await this.getCredentials('proboApi');

	if (!credentials!.apiKey) {
		throw new Error('API Key is required');
	}

	const responseData = await this.helpers.httpRequest({
		method: 'POST',
		baseURL: `${credentials!.server}`,
		url: '/api/console/v1/query',
		headers: {
			Authorization: `Bearer ${credentials!.apiKey}`,
			'Content-Type': 'application/json',
		},
		body: body,
		json: true,	
		returnFullResponse: true,
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

