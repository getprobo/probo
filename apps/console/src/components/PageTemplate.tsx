import type { FC, ReactNode } from "react";
import {
  PageContainer,
  type PageContainerProps,
  PageContainerSkeleton,
} from "./PageContainer";
import { PageHeader, type PageHeaderProps, PageHeaderSkeleton } from "./PageHeader";

export const PageTemplate: FC<PageContainerProps & PageHeaderProps> = ({
  actions,
  title,
  description,
  children,
}) => {
  return (
    <PageContainer title={title}>
      <PageHeader
        className="mb-12"
        title={title}
        description={description}
        actions={actions}
      />
      {children}
    </PageContainer>
  );
};

export const PageTemplateSkeleton: FC<{
  title?: string;
  description?: ReactNode;
  withDescription?: boolean;
  actions?: ReactNode;
  children: ReactNode;
}> = ({ actions, children, description, withDescription = false, title }) => {
  return (
    <PageContainerSkeleton title={title}>
      <PageHeaderSkeleton
        className="mb-17"
        title={title}
        description={description}
        withDescription={withDescription}
        actions={actions}
      />
      {children}
    </PageContainerSkeleton>
  );
};
