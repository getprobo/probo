import { PageTemplateSkeleton } from "@/components/PageTemplate";
import { Suspense } from "react";
import { lazy } from "@probo/react-lazy";
import { useLocation } from "react-router";
import { ErrorBoundaryWithLocation } from "../ErrorBoundary";

const CreatePeopleView = lazy(() => import("./CreatePeopleView"));

export function CreatePeopleViewSkeleton() {
	return (
		<PageTemplateSkeleton
			title="Create Person"
			description="Add a new person interacting with organization"
		>
			<div className="max-w-2xl aspect-square bg-muted rounded-xl animate-pulse" />
		</PageTemplateSkeleton>
	);
}

export function CreatePeoplePage() {
	const location = useLocation();

	return (
		<Suspense key={location.pathname} fallback={<CreatePeopleViewSkeleton />}>
			<ErrorBoundaryWithLocation>
				<CreatePeopleView />
			</ErrorBoundaryWithLocation>
		</Suspense>
	);
}
