import {
	IAuthenticateGeneric,
	ICredentialTestRequest,
	ICredentialType,
	Icon,
	INodeProperties,
} from 'n8n-workflow';

export class ProboApi implements ICredentialType {
	name = 'proboApi';

	displayName = 'Probo API';

	icon: Icon = { light: 'file:../icons/probo-light.svg', dark: 'file:../icons/probo.svg' };

	documentationUrl = 'https://www.getprobo.com/docs';

	properties: INodeProperties[] = [
		{
			displayName: 'Probo Server',
			name: 'server',
			type: 'string',
			default: 'https://us.console.getprobo.com',
			description: 'The server to connect to.',
		},
		{
			displayName: 'API Key',
			name: 'apiKey',
			type: 'string',
			typeOptions: { password: true },
			default: '',
		},
	];

	authenticate: IAuthenticateGeneric = {
		type: 'generic',
		properties: {
			headers: {
				Authorization: '=Bearer {{$credentials?.apiKey}}',
			},
		},
	};

	test: ICredentialTestRequest = {
		request: {
			baseURL: '={{$credentials?.server}}',
			url: '/api/console/v1/graphql',
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			body: {
				query: 'query { viewer { id } }',
			},
		},
	};
}
