# Fix Incident Registry Implementation

## Summary
Complete the incident registry implementation by integrating the backend services with the frontend GraphQL layer and replacing mock data with real functionality.

## Background
The incident registry feature has been partially implemented with:
- ✅ Complete backend schema and services
- ✅ Frontend pages and components with mock data
- ✅ Routing and sidebar navigation
- ✅ Create dialog with form validation
- ❌ Missing GraphQL integration layer
- ❌ Mock data instead of real backend calls

## Tasks

### Backend Integration
- [ ] **Complete GraphQL Schema**: Add remaining GraphQL types and mutations to `pkg/server/api/console/v1/schema.graphql`
  - [ ] Add `IncidentRegistry` type definition
  - [ ] Add `IncidentRegistryConnection` and `IncidentRegistryEdge` types  
  - [ ] Add mutations: `createIncidentRegistry`, `updateIncidentRegistry`, `deleteIncidentRegistry`
  - [ ] Add input types: `CreateIncidentRegistryInput`, `UpdateIncidentRegistryInput`, `DeleteIncidentRegistryInput`
  - [ ] Add payload types for mutations

- [ ] **GraphQL Resolvers**: Implement resolvers that call the backend services
  - [ ] `Organization.incidentRegistries` connection resolver
  - [ ] `createIncidentRegistry` mutation resolver
  - [ ] `updateIncidentRegistry` mutation resolver  
  - [ ] `deleteIncidentRegistry` mutation resolver

### Frontend Integration
- [ ] **GraphQL Hooks**: Create React hooks for incident registry operations
  - [ ] Create `apps/console/src/hooks/graph/IncidentRegistryGraph.ts`
  - [ ] Implement `useIncidentRegistries` query hook
  - [ ] Implement `useIncidentRegistry` single item hook
  - [ ] Implement `useCreateIncidentRegistry` mutation hook
  - [ ] Implement `useUpdateIncidentRegistry` mutation hook
  - [ ] Implement `useDeleteIncidentRegistry` mutation hook

- [ ] **Replace Mock Data**: Update components to use real GraphQL data
  - [ ] Replace mock data in `IncidentRegistriesPage.tsx` with `useIncidentRegistries` hook
  - [ ] Replace mock data in `IncidentRegistryDetailsPage.tsx` with `useIncidentRegistry` hook
  - [ ] Connect `CreateIncidentRegistryDialog.tsx` to real mutation
  - [ ] Add proper connection/pagination support

- [ ] **Add Missing Features**: Complete CRUD functionality
  - [ ] Add edit/update dialog or inline editing
  - [ ] Implement delete functionality with confirmation
  - [ ] Add proper error handling and loading states
  - [ ] Add real-time updates and optimistic updates

### Testing & Polish
- [ ] **Database Migration**: Ensure migration runs correctly
  - [ ] Test `pkg/coredata/migrations/20250828T150000Z.sql`
  - [ ] Verify all indexes and constraints work properly

- [ ] **End-to-End Testing**: Test complete user workflows
  - [ ] Create incident → verify in database
  - [ ] Edit incident → verify updates persist  
  - [ ] Delete incident → verify removal
  - [ ] Test pagination and filtering
  - [ ] Test form validation and error scenarios

- [ ] **Performance**: Optimize queries and loading
  - [ ] Add proper GraphQL field selections
  - [ ] Implement pagination cursors correctly
  - [ ] Add loading skeletons where appropriate

## Files to Modify

### Backend
- `pkg/server/api/console/v1/schema.graphql` - Add GraphQL types and mutations
- `pkg/server/api/console/v1/schema/` - Add resolver implementations (if using separate files)

### Frontend  
- `apps/console/src/hooks/graph/IncidentRegistryGraph.ts` - **NEW FILE** - GraphQL hooks
- `apps/console/src/pages/organizations/incidentRegistries/IncidentRegistriesPage.tsx` - Replace mock data
- `apps/console/src/pages/organizations/incidentRegistries/IncidentRegistryDetailsPage.tsx` - Replace mock data
- `apps/console/src/pages/organizations/incidentRegistries/dialogs/CreateIncidentRegistryDialog.tsx` - Connect to real mutation

## Definition of Done
- [ ] All CRUD operations work end-to-end (Create, Read, Update, Delete)
- [ ] No mock data remains in the codebase
- [ ] Database operations persist correctly
- [ ] Form validation works with backend constraints
- [ ] Error handling provides meaningful user feedback
- [ ] Loading states provide good user experience
- [ ] All TypeScript types are properly generated
- [ ] Component tests pass (if applicable)

## Priority
**High** - This completes a major feature and removes technical debt from mock implementations.

## Estimated Effort
**Medium (4-6 hours)** - Involves both backend and frontend integration work.

## Notes
- The backend service layer (`pkg/probo/incident_registry_service.go`) is already complete
- The database schema and migration are ready
- The frontend UI components are fully styled and functional
- This is primarily an integration task to connect existing pieces