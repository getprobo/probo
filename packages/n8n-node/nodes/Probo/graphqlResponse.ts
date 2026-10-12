// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import type { JsonObject as JSONObject } from 'n8n-workflow';

export type GraphQLBodyResult =
	| { ok: true; body: JSONObject }
	| { ok: false; message: string; body?: JSONObject };

function isJSONObject(value: unknown): value is JSONObject {
	return typeof value === 'object' && value !== null && !Array.isArray(value) && !Buffer.isBuffer(value);
}

function unwrapHTTPEnvelope(value: JSONObject): JSONObject | undefined {
	if (!('statusCode' in value) || !('body' in value)) {
		return value;
	}

	const body = value.body;
	if (typeof body === 'string' || Buffer.isBuffer(body)) {
		const text = Buffer.isBuffer(body) ? body.toString('utf8') : body;
		try {
			const parsed: unknown = JSON.parse(text);
			return isJSONObject(parsed) ? parsed : undefined;
		} catch {
			return undefined;
		}
	}

	return isJSONObject(body) ? body : undefined;
}

function graphqlErrorMessage(body: JSONObject): string | undefined {
	const errors = body.errors;
	if (!Array.isArray(errors) || errors.length === 0) {
		return undefined;
	}

	return errors
		.map((err) => {
			if (isJSONObject(err) && typeof err.message === 'string' && err.message !== '') {
				return err.message;
			}
			return JSON.stringify(err);
		})
		.join('; ');
}

export function readGraphQLBody(response: unknown): GraphQLBodyResult {
	let body: JSONObject | undefined;
	if (typeof response === 'string' || Buffer.isBuffer(response)) {
		const text = Buffer.isBuffer(response) ? response.toString('utf8') : response;
		try {
			const parsed: unknown = JSON.parse(text);
			body = isJSONObject(parsed) ? parsed : undefined;
		} catch {
			body = undefined;
		}
	} else if (isJSONObject(response)) {
		body = unwrapHTTPEnvelope(response);
	}

	if (body === undefined) {
		return { ok: false, message: 'Probo returned a response that is not JSON' };
	}

	const message = graphqlErrorMessage(body);
	if (message !== undefined) {
		return { ok: false, message, body };
	}

	return { ok: true, body };
}
