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

import type { IDataObject, ILoadOptionsFunctions, INodePropertyOptions } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

const connectorProvidersQuery = `
	query ConnectorProviders {
		connectorProviders {
			provider
			displayName
			apiKeySupported
			apiKeyManaged
			installSupported
			apiKeyExtraSettings {
				key
			}
			clientCredentialsSupported
			clientCredentialsExtraSettings {
				key
			}
		}
	}
`;

interface CatalogProvider {
	provider: string;
	displayName: string;
	apiKeySupported: boolean;
	apiKeyManaged: boolean;
	installSupported: boolean;
	apiKeyExtraSettings: Array<{ key: string }>;
	clientCredentialsSupported: boolean;
	clientCredentialsExtraSettings: Array<{ key: string }>;
}

export async function getAPIKeyConnectorProviders(
	this: ILoadOptionsFunctions,
): Promise<INodePropertyOptions[]> {
	const providers = await loadConnectorProviders.call(this);

	return providers
		.filter(
			(provider) =>
				provider.apiKeySupported &&
				!provider.apiKeyManaged &&
				!provider.installSupported &&
				provider.apiKeyExtraSettings.length === 0,
		)
		.map(toOption);
}

export async function getClientCredentialsConnectorProviders(
	this: ILoadOptionsFunctions,
): Promise<INodePropertyOptions[]> {
	const providers = await loadConnectorProviders.call(this);

	return providers
		.filter(
			(provider) =>
				provider.clientCredentialsSupported &&
				provider.clientCredentialsExtraSettings.length === 0,
		)
		.map(toOption);
}

async function loadConnectorProviders(this: ILoadOptionsFunctions): Promise<CatalogProvider[]> {
	const responseData = await proboApiRequest.call(this, connectorProvidersQuery, {});
	const data = responseData.data as IDataObject | undefined;
	const providers = data?.connectorProviders;

	if (!Array.isArray(providers)) {
		return [];
	}

	return providers as CatalogProvider[];
}

function toOption(provider: CatalogProvider): INodePropertyOptions {
	return {
		name: provider.displayName,
		value: provider.provider,
	};
}
