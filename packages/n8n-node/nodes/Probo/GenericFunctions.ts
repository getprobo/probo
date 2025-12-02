import type {
	IExecuteFunctions,
	IHookFunctions,
	IDataObject,
	JsonObject,
	IHttpRequestOptions,
} from 'n8n-workflow';
import { NodeApiError } from 'n8n-workflow';

export async function proboApiRequest(
	this: IExecuteFunctions | IHookFunctions,
	query: string,
	variables: IDataObject = {},
): Promise<any> {
	const credentials = await this.getCredentials('proboApi');

	if (!credentials?.apiKey) {
		throw new NodeApiError(this.getNode(), { message: 'API Key is required' } as JsonObject);
	}

	const options: IHttpRequestOptions = {
		method: 'POST',
		baseURL: `${credentials.server}`,
		url: '/api/console/v1/query',
		headers: {
			Authorization: `Bearer ${credentials.apiKey}`,
			'Content-Type': 'application/json',
		},
		body: {
			query,
			variables,
		},
		json: true,
	};

	try {
		return await this.helpers.httpRequest(options);
	} catch (error) {
		throw new NodeApiError(this.getNode(), error as JsonObject);
	}
}

export async function proboApiRequestAllItems(
	this: IExecuteFunctions,
	query: string,
	variables: IDataObject,
	getConnection: (response: any) => any,
	returnAll: boolean = true,
	limit: number = 0,
): Promise<any[]> {
	const items: any[] = [];
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

		const responseData = await proboApiRequest.call(this, query, requestVariables);
		const connection = getConnection(responseData);

		if (connection?.edges) {
			items.push(...connection.edges.map((edge: any) => edge.node));
		}

		if (connection?.pageInfo) {
			hasNextPage = connection.pageInfo.hasNextPage;
			cursor = connection.pageInfo.endCursor;
		} else {
			hasNextPage = false;
		}

		if (!returnAll && items.length >= limit) {
			hasNextPage = false;
		}
	}

	return items;
}
