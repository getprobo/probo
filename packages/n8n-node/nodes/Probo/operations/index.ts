import type { IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { executeGraphQL } from './execute';

export async function executeOperation(
	this: IExecuteFunctions,
	operation: string,
	itemIndex: number,
): Promise<INodeExecutionData> {
	switch (operation) {
		case 'execute':
			return executeGraphQL.call(this, itemIndex);
		default:
			throw new Error(`Unknown operation: ${operation}`);
	}
}

