import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Risk ID',
		name: 'riskId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['linkDocument'],
			},
		},
		default: '',
		description: 'The ID of the risk',
		required: true,
	},
	{
		displayName: 'Document ID',
		name: 'documentId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['linkDocument'],
			},
		},
		default: '',
		description: 'The ID of the document to link',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskId = this.getNodeParameter('riskId', itemIndex) as string;
	const documentId = this.getNodeParameter('documentId', itemIndex) as string;

	const query = `
		mutation CreateRiskDocumentMapping($input: CreateRiskDocumentMappingInput!) {
			createRiskDocumentMapping(input: $input) {
				riskEdge {
					node {
						id
						name
					}
				}
				documentEdge {
					node {
						id
						title
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { riskId, documentId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
