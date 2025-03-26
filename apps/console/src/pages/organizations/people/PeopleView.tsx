"use client";

import { Card } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import { HelpCircle } from "lucide-react";
import {
	graphql,
	PreloadedQuery,
	usePreloadedQuery,
	useQueryLoader,
	useMutation,
} from "react-relay";
import { Suspense, useEffect, useState, useCallback } from "react";
import type { PeopleOverviewPageQuery as PeopleOverviewPageQueryType } from "./__generated__/PeopleOverviewPageQuery.graphql";
import { useParams } from "react-router";
import { Helmet } from "react-helmet-async";
import { cn } from "@/lib/utils";

const peopleOverviewPageQuery = graphql`
  query PeopleOverviewPageQuery($peopleId: ID!) {
    node(id: $peopleId) {
      ... on People {
        id
        fullName
        primaryEmailAddress
        additionalEmailAddresses
        kind
        position
        createdAt
        updatedAt
        version
      }
    }
  }
`;

const updatePeopleMutation = graphql`
  mutation PeopleOverviewPageUpdatePeopleMutation($input: UpdatePeopleInput!) {
    updatePeople(input: $input) {
      people {
        id
        fullName
        primaryEmailAddress
        additionalEmailAddresses
        kind
        position
        updatedAt
        version
      }
    }
  }
`;

function EditableField({
	label,
	value,
	onChange,
	type = "text",
	helpText,
}: {
	label: string;
	value: string;
	onChange: (value: string) => void;
	type?: string;
	helpText?: string;
}) {
	return (
		<div className="space-y-2">
			<div className="flex items-center gap-2">
				<HelpCircle className="h-4 w-4 text-gray-400" />
				<Label className="text-sm">{label}</Label>
			</div>
			<div className="space-y-2">
				<Input
					type={type}
					value={value}
					onChange={(e) => onChange(e.target.value)}
				/>
				{helpText && <p className="text-sm text-gray-500">{helpText}</p>}
			</div>
		</div>
	);
}

function PeopleOverviewPageContent({
	queryRef,
}: {
	queryRef: PreloadedQuery<PeopleOverviewPageQueryType>;
}) {
	const data = usePreloadedQuery(peopleOverviewPageQuery, queryRef);
	const [editedFields, setEditedFields] = useState<Set<string>>(new Set());
	const [formData, setFormData] = useState({
		fullName: data.node.fullName || "",
		primaryEmailAddress: data.node.primaryEmailAddress || "",
		additionalEmailAddresses: data.node.additionalEmailAddresses || [],
		kind: data.node.kind,
	});
	const [commit] = useMutation(updatePeopleMutation);
	const [, loadQuery] = useQueryLoader<PeopleOverviewPageQueryType>(
		peopleOverviewPageQuery,
	);
	const { toast } = useToast();

	const hasChanges = editedFields.size > 0;

	const handleSave = useCallback(() => {
		commit({
			variables: {
				input: {
					id: data.node.id,
					expectedVersion: data.node.version,
					...formData,
				},
			},
			onCompleted: () => {
				toast({
					title: "Success",
					description: "Changes saved successfully",
					variant: "default",
				});
				setEditedFields(new Set());
			},
			onError: (error) => {
				if (error.message?.includes("concurrent modification")) {
					toast({
						title: "Error",
						description:
							"Someone else modified this person. Reloading latest data.",
						variant: "destructive",
					});
					loadQuery({ peopleId: data.node.id! });
				} else {
					toast({
						title: "Error",
						description: error.message || "Failed to save changes",
						variant: "destructive",
					});
				}
			},
		});
	}, [commit, data.node.id, data.node.version, formData, loadQuery, toast]);

	const handleFieldChange = (field: keyof typeof formData, value: unknown) => {
		setFormData((prev) => ({
			...prev,
			[field]: value,
		}));
		setEditedFields((prev) => new Set(prev).add(field));
	};

	const handleCancel = () => {
		setFormData({
			fullName: data.node.fullName || "",
			primaryEmailAddress: data.node.primaryEmailAddress || "",
			additionalEmailAddresses: data.node.additionalEmailAddresses || [],
			kind: data.node.kind,
		});
		setEditedFields(new Set());
	};

	return (
		<>
			<div className="space-y-6 p-4 md:p-6 lg:p-8">
				<div className="mx-auto max-w-4xl space-y-6">
					<EditableField
						label="Full Name"
						value={formData.fullName}
						onChange={(value) => handleFieldChange("fullName", value)}
					/>

					<EditableField
						label="Primary Email"
						value={formData.primaryEmailAddress}
						type="email"
						onChange={(value) =>
							handleFieldChange("primaryEmailAddress", value)
						}
					/>

					<div className="space-y-2">
						<div className="flex items-center gap-2">
							<HelpCircle className="h-4 w-4 text-gray-400" />
							<Label className="text-sm">Additional Email Addresses</Label>
						</div>
						<div className="space-y-2">
							{formData.additionalEmailAddresses.map((email, index) => (
								<div key={index} className="flex gap-2">
									<Input
										type="email"
										value={email}
										onChange={(e) => {
											const newEmails = [...formData.additionalEmailAddresses];
											newEmails[index] = e.target.value;
											handleFieldChange("additionalEmailAddresses", newEmails);
										}}
									/>
									<Button
										variant="outline"
										onClick={() => {
											const newEmails =
												formData.additionalEmailAddresses.filter(
													(_, i) => i !== index,
												);
											handleFieldChange("additionalEmailAddresses", newEmails);
										}}
									>
										Remove
									</Button>
								</div>
							))}
							<Button
								variant="outline"
								onClick={() => {
									handleFieldChange("additionalEmailAddresses", [
										...formData.additionalEmailAddresses,
										"",
									]);
								}}
							>
								Add Email
							</Button>
						</div>
					</div>

					<Card className="p-6">
						<div className="space-y-4">
							<div className="space-y-2">
								<h2 className="text-lg font-medium">Additional Information</h2>
								<p className="text-sm text-gray-500">
									Additional details about the person
								</p>
							</div>

							<div className="space-y-4">
								<div className="space-y-2">
									<div className="flex items-center gap-2">
										<HelpCircle className="h-4 w-4 text-gray-400" />
										<Label className="text-sm">Kind</Label>
									</div>
									<div className="flex gap-2">
										<button
											onClick={() => handleFieldChange("kind", "EMPLOYEE")}
											className={cn(
												"rounded-full px-4 py-1 text-sm transition-colors",
												formData.kind === "EMPLOYEE"
													? "bg-blue-100 text-blue-900 ring-2 ring-blue-600 ring-offset-2"
													: "bg-gray-100 text-gray-900 hover:bg-gray-200",
											)}
										>
											Employee
										</button>
										<button
											onClick={() => handleFieldChange("kind", "CONTRACTOR")}
											className={cn(
												"rounded-full px-4 py-1 text-sm transition-colors",
												formData.kind === "CONTRACTOR"
													? "bg-purple-100 text-purple-900 ring-2 ring-purple-600 ring-offset-2"
													: "bg-gray-100 text-gray-900 hover:bg-gray-200",
											)}
										>
											Contractor
										</button>
										<button
											onClick={() =>
												handleFieldChange("kind", "SERVICE_ACCOUNT")
											}
											className={cn(
												"rounded-full px-4 py-1 text-sm transition-colors",
												formData.kind === "SERVICE_ACCOUNT"
													? "bg-green-100 text-green-900 ring-2 ring-green-600 ring-offset-2"
													: "bg-gray-100 text-gray-900 hover:bg-gray-200",
											)}
										>
											Service Account
										</button>
									</div>
								</div>
							</div>
						</div>
					</Card>
				</div>
			</div>

			{hasChanges && (
				<div className="fixed bottom-6 right-6 flex gap-2">
					<Button variant="outline" onClick={handleCancel}>
						Cancel
					</Button>
					<Button
						onClick={handleSave}
						className="bg-primary text-primary-foreground hover:bg-primary/90"
					>
						Save Changes
					</Button>
				</div>
			)}
		</>
	);
}

function PeopleOverviewPageFallback() {
	return (
		<div className="p-6 space-y-6">
			<div className="space-y-1">
				<div className="h-8 w-48 bg-muted animate-pulse rounded" />
				<div className="h-4 w-96 bg-muted animate-pulse rounded" />
			</div>
			<div className="space-y-2">
				{[1, 2].map((i) => (
					<div key={i} className="h-20 bg-muted animate-pulse rounded-lg" />
				))}
			</div>
		</div>
	);
}

export default function PeopleOverviewPage() {
	const { peopleId } = useParams();
	const [queryRef, loadQuery] = useQueryLoader<PeopleOverviewPageQueryType>(
		peopleOverviewPageQuery,
	);

	useEffect(() => {
		loadQuery({ peopleId: peopleId! });
	}, [loadQuery, peopleId]);

	if (!queryRef) {
		return <PeopleOverviewPageFallback />;
	}

	return (
		<>
			<Helmet>
				<title>People Overview - Probo Console</title>
			</Helmet>
			<Suspense fallback={<PeopleOverviewPageFallback />}>
				<PeopleOverviewPageContent queryRef={queryRef} />
			</Suspense>
		</>
	);
}
