import type {
	IExecuteFunctions,
	IHookFunctions,
	IDataObject,
	JsonObject,
	IHttpRequestOptions,
} from 'n8n-workflow';
import { NodeApiError } from 'n8n-workflow';

import { version } from '../../package.json';

type ApiRequestFn = (
	this: IExecuteFunctions | IHookFunctions,
	query: string,
	variables?: IDataObject,
) => Promise<IDataObject>;

async function proboGraphqlRequest(
	this: IExecuteFunctions | IHookFunctions,
	apiPath: string,
	query: string,
	variables: IDataObject = {},
): Promise<IDataObject> {
	const credentials = await this.getCredentials('proboApi');

	if (!credentials?.apiKey) {
		throw new NodeApiError(this.getNode(), { message: 'API Key is required' } as JsonObject);
	}

	const options: IHttpRequestOptions = {
		method: 'POST',
		baseURL: `${credentials.server}`,
		url: apiPath,
		headers: {
			Authorization: `Bearer ${credentials.apiKey}`,
			'Content-Type': 'application/json',
			'User-Agent': `probo-n8n-node/${version}`,
		},
		body: {
			query,
			variables,
		},
		json: true,
	};

	try {
		const response = await this.helpers.httpRequest(options);

		if (response.errors && Array.isArray(response.errors) && response.errors.length > 0) {
			const errorMessages = response.errors.map((err: IDataObject) =>
				err.message || JSON.stringify(err)
			).join('; ');
			throw new NodeApiError(this.getNode(), {
				message: `GraphQL errors: ${errorMessages}`,
				httpCode: '200',
			} as JsonObject);
		}

		return response;
	} catch (error) {
		throw new NodeApiError(this.getNode(), error as JsonObject);
	}
}

export async function proboApiRequest(
	this: IExecuteFunctions | IHookFunctions,
	query: string,
	variables: IDataObject = {},
): Promise<IDataObject> {
	return proboGraphqlRequest.call(this, '/api/console/v1/graphql', query, variables);
}

export async function proboConnectApiRequest(
	this: IExecuteFunctions | IHookFunctions,
	query: string,
	variables: IDataObject = {},
): Promise<IDataObject> {
	return proboGraphqlRequest.call(this, '/api/connect/v1/graphql', query, variables);
}

async function proboGraphqlRequestAllItems(
	this: IExecuteFunctions,
	requestFn: ApiRequestFn,
	query: string,
	variables: IDataObject,
	getConnection: (response: IDataObject) => IDataObject | undefined,
	returnAll: boolean = true,
	limit: number = 0,
): Promise<IDataObject[]> {
	const items: IDataObject[] = [];
	let hasNextPage = true;
	let cursor: string | null = null;
	const pageSize = 100;

	while (hasNextPage) {
		const currentLimit = returnAll ? pageSize : Math.min(pageSize, limit - items.length);

		if (currentLimit <= 0) {
			break;
		}

		const requestVariables: IDataObject = {
			...variables,
			first: currentLimit,
		};
		if (cursor) {
			requestVariables.after = cursor;
		}

		const responseData = await requestFn.call(this, query, requestVariables);
		const connection = getConnection(responseData);

		if (connection?.edges) {
			const edges = connection.edges as Array<{ node: IDataObject }>;
			items.push(...edges.map((edge) => edge.node));
		}

		if (connection?.pageInfo) {
			const pageInfo = connection.pageInfo as IDataObject;
			hasNextPage = pageInfo.hasNextPage as boolean;
			cursor = pageInfo.endCursor as string | null;
		} else {
			hasNextPage = false;
		}

		if (!returnAll && items.length >= limit) {
			hasNextPage = false;
		}
	}

	return items;
}

export async function proboApiRequestAllItems(
	this: IExecuteFunctions,
	query: string,
	variables: IDataObject,
	getConnection: (response: IDataObject) => IDataObject | undefined,
	returnAll: boolean = true,
	limit: number = 0,
): Promise<IDataObject[]> {
	return proboGraphqlRequestAllItems.call(
		this,
		proboApiRequest,
		query,
		variables,
		getConnection,
		returnAll,
		limit,
	);
}

export async function proboConnectApiRequestAllItems(
	this: IExecuteFunctions,
	query: string,
	variables: IDataObject,
	getConnection: (response: IDataObject) => IDataObject | undefined,
	returnAll: boolean = true,
	limit: number = 0,
): Promise<IDataObject[]> {
	return proboGraphqlRequestAllItems.call(
		this,
		proboConnectApiRequest,
		query,
		variables,
		getConnection,
		returnAll,
		limit,
	);
}
