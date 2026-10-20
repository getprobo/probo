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
import { NodeApiError, NodeOperationError } from 'n8n-workflow';

import { version } from '../../package.json';
import { readGraphQLBody } from './graphqlResponse';

type ApiRequestFn = (
	this: IExecuteFunctions | IHookFunctions,
	query: string,
	variables?: IDataObject,
) => Promise<IDataObject>;

type ApiRequestContext = IExecuteFunctions | IHookFunctions | ILoadOptionsFunctions;

function graphqlBodyOrThrow(
	this: ApiRequestContext,
	response: unknown,
	options?: { keepPartialResult?: boolean },
): IDataObject {
	const parsed = readGraphQLBody(response);
	if (!parsed.ok) {
		if (options?.keepPartialResult && parsed.body?.data != null) {
			return parsed.body as IDataObject;
		}

		throw new NodeApiError(this.getNode(), (parsed.body ?? { message: parsed.message }) as JsonObject, {
			message: parsed.message,
			level: 'error',
		});
	}

	return parsed.body as IDataObject;
}

function rethrowRequestError(this: ApiRequestContext, error: unknown): never {
	if (error instanceof NodeApiError || error instanceof NodeOperationError) {
		throw error;
	}

	throw new NodeApiError(this.getNode(), error as JsonObject);
}

async function proboGraphQLRequest(
	this: ApiRequestContext,
	apiPath: string,
	query: string,
	variables: IDataObject = {},
	resultOptions?: { keepPartialResult?: boolean },
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

	let response: unknown;
	try {
		response = await this.helpers.httpRequestWithAuthentication.call(
			this,
			'proboApi',
			options,
		);
	} catch (error) {
		rethrowRequestError.call(this, error);
	}

	return graphqlBodyOrThrow.call(this, response, resultOptions);
}

export async function proboApiRequest(
	this: ApiRequestContext,
	query: string,
	variables: IDataObject = {},
	options?: { keepPartialResult?: boolean },
): Promise<IDataObject> {
	return proboGraphQLRequest.call(this, '/api/console/v1/graphql', query, variables, options);
}

export async function proboConnectApiRequest(
	this: IExecuteFunctions | IHookFunctions,
	query: string,
	variables: IDataObject = {},
): Promise<IDataObject> {
	return proboGraphQLRequest.call(this, '/api/connect/v1/graphql', query, variables);
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

	let response: unknown;
	try {
		response = await this.helpers.httpRequestWithAuthentication.call(
			this,
			'proboApi',
			options,
		);
	} catch (error) {
		rethrowRequestError.call(this, error);
	}

	return graphqlBodyOrThrow.call(this, response);
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
