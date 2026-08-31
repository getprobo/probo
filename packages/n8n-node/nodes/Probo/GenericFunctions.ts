// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import type {
	IExecuteFunctions,
	IHookFunctions,
	ILoadOptionsFunctions,
	IDataObject,
	JsonObject,
	IHttpRequestOptions,
} from 'n8n-workflow';
import { NodeApiError } from 'n8n-workflow';

import { version } from '../../package.json';

function graphqlErrorMessages(errors: IDataObject[]): string {
	return errors
		.map((err) => (typeof err.message === 'string' && err.message) || JSON.stringify(err))
		.join('; ');
}

function graphqlErrorsFromUnknown(
	value: unknown,
	seen: Set<unknown> = new Set(),
): IDataObject[] | undefined {
	if (!value || typeof value !== 'object' || seen.has(value)) {
		return undefined;
	}

	seen.add(value);

	const record = value as IDataObject;
	if (Array.isArray(record.errors)) {
		return record.errors as IDataObject[];
	}

	for (const key of ['data', 'body', 'response', 'error', 'cause'] as const) {
		const nested = graphqlErrorsFromUnknown(record[key], seen);
		if (nested) {
			return nested;
		}
	}

	return undefined;
}

function graphqlErrorMessagesFromHttpError(error: unknown): string | undefined {
	const errors = graphqlErrorsFromUnknown(error);
	if (!errors || errors.length === 0) {
		return undefined;
	}

	return graphqlErrorMessages(errors);
}

function httpCodeFromError(error: unknown): string | undefined {
	if (!error || typeof error !== 'object') {
		return undefined;
	}

	const record = error as IDataObject;
	if (typeof record.httpCode === 'string' || typeof record.httpCode === 'number') {
		return String(record.httpCode);
	}

	if (typeof record.status === 'number' || typeof record.status === 'string') {
		return String(record.status);
	}

	const response = record.response as IDataObject | undefined;
	if (typeof response?.status === 'number' || typeof response?.status === 'string') {
		return String(response.status);
	}

	if (typeof response?.statusCode === 'number' || typeof response?.statusCode === 'string') {
		return String(response.statusCode);
	}

	return undefined;
}

type ApiRequestFn = (
	this: IExecuteFunctions | IHookFunctions,
	query: string,
	variables?: IDataObject,
) => Promise<IDataObject>;

type ApiRequestContext = IExecuteFunctions | IHookFunctions | ILoadOptionsFunctions;

async function proboGraphqlRequest(
	this: ApiRequestContext,
	apiPath: string,
	query: string,
	variables: IDataObject = {},
): Promise<IDataObject> {
	const credentials = await this.getCredentials('proboApi');

	const options: IHttpRequestOptions = {
		method: 'POST',
		baseURL: `${credentials.server}`,
		url: apiPath,
		headers: {
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
		const response = await this.helpers.httpRequestWithAuthentication.call(
			this,
			'proboApi',
			options,
		);

		if (response.errors && Array.isArray(response.errors) && response.errors.length > 0) {
			throw new NodeApiError(this.getNode(), {
				message: `GraphQL errors: ${graphqlErrorMessages(response.errors)}`,
				httpCode: '200',
			} as JsonObject);
		}

		return response;
	} catch (error) {
		const graphqlMessage = graphqlErrorMessagesFromHttpError(error);
		if (graphqlMessage) {
			throw new NodeApiError(this.getNode(), {
				message: `GraphQL errors: ${graphqlMessage}`,
				httpCode: httpCodeFromError(error) ?? '422',
			} as JsonObject);
		}

		throw new NodeApiError(this.getNode(), error as JsonObject);
	}
}

export async function proboApiRequest(
	this: ApiRequestContext,
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
		proboApiRequest as ApiRequestFn,
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
		proboConnectApiRequest as ApiRequestFn,
		query,
		variables,
		getConnection,
		returnAll,
		limit,
	);
}

export async function proboApiMultipartRequest(
	this: IExecuteFunctions,
	query: string,
	variables: IDataObject,
	fileVariablePath: string,
	fileBuffer: Buffer,
	fileName: string,
	mimeType: string = 'application/octet-stream',
): Promise<IDataObject> {
	const credentials = await this.getCredentials('proboApi');

	const boundary = `----n8nFormBoundary${Date.now().toString(16)}`;

	const safeFileName = fileName
		.replace(/[\r\n]/g, '')
		.replace(/\\/g, '\\\\')
		.replace(/"/g, '\\"');
	const safeMimeType = mimeType.replace(/[\r\n]/g, '');

	const operations = JSON.stringify({ query, variables });
	const map = JSON.stringify({ '0': [fileVariablePath] });

	const parts: Buffer[] = [];

	parts.push(Buffer.from(
		`--${boundary}\r\nContent-Disposition: form-data; name="operations"\r\n\r\n${operations}\r\n`,
	));

	parts.push(Buffer.from(
		`--${boundary}\r\nContent-Disposition: form-data; name="map"\r\n\r\n${map}\r\n`,
	));

	parts.push(Buffer.from(
		`--${boundary}\r\nContent-Disposition: form-data; name="0"; filename="${safeFileName}"\r\nContent-Type: ${safeMimeType}\r\n\r\n`,
	));
	parts.push(fileBuffer);
	parts.push(Buffer.from(`\r\n--${boundary}--\r\n`));

	const body = Buffer.concat(parts);

	const options: IHttpRequestOptions = {
		method: 'POST',
		baseURL: `${credentials.server}`,
		url: '/api/console/v1/graphql',
		headers: {
			'Content-Type': `multipart/form-data; boundary=${boundary}`,
			'User-Agent': `probo-n8n-node/${version}`,
		},
		body,
	};

	try {
		const response = await this.helpers.httpRequestWithAuthentication.call(
			this,
			'proboApi',
			options,
		);

		if (response.errors && Array.isArray(response.errors) && response.errors.length > 0) {
			throw new NodeApiError(this.getNode(), {
				message: `GraphQL errors: ${graphqlErrorMessages(response.errors)}`,
				httpCode: '200',
			} as JsonObject);
		}

		return response;
	} catch (error) {
		const graphqlMessage = graphqlErrorMessagesFromHttpError(error);
		if (graphqlMessage) {
			throw new NodeApiError(this.getNode(), {
				message: `GraphQL errors: ${graphqlMessage}`,
				httpCode: httpCodeFromError(error) ?? '422',
			} as JsonObject);
		}

		throw new NodeApiError(this.getNode(), error as JsonObject);
	}
}

export function toPeriod(
	start?: string | null,
	end?: string | null,
): { start?: string; end?: string } | undefined {
	if (!start && !end) {
		return undefined;
	}

	const period: { start?: string; end?: string } = {};
	if (start) {
		period.start = start;
	}

	if (end) {
		period.end = end;
	}

	return period;
}
