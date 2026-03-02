import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiMultipartRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Audit ID',
		name: 'auditId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['audit'],
				operation: ['uploadReport'],
			},
		},
		default: '',
		description: 'The ID of the audit to upload the report for',
		required: true,
	},
	{
		displayName: 'Input Data Field Name',
		name: 'binaryPropertyName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['audit'],
				operation: ['uploadReport'],
			},
		},
		default: 'data',
		description: 'The name of the input field containing the binary file data to upload',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const auditId = this.getNodeParameter('auditId', itemIndex) as string;
	const binaryPropertyName = this.getNodeParameter('binaryPropertyName', itemIndex) as string;

	const binaryData = this.helpers.assertBinaryData(itemIndex, binaryPropertyName);
	const fileBuffer = await this.helpers.getBinaryDataBuffer(itemIndex, binaryPropertyName);

	const fileName = binaryData.fileName || 'report';
	const mimeType = binaryData.mimeType || 'application/octet-stream';

	const query = `
		mutation UploadAuditReport($input: UploadAuditReportInput!) {
			uploadAuditReport(input: $input) {
				audit {
					id
					name
					state
					validFrom
					validUntil
					reportUrl
					trustCenterVisibility
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		input: {
			auditId,
			file: null,
		},
	};

	const responseData = await proboApiMultipartRequest.call(
		this,
		query,
		variables,
		'variables.input.file',
		fileBuffer,
		fileName,
		mimeType,
	);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
