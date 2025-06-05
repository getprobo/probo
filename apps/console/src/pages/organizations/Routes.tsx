import { Route, Routes } from "react-router";
import NotFoundPage from "../NotFoundPage";
import { CreateOrganizationPage } from "./CreateOrganizationPage";
import HomePage from "./HomePage";
import NoOrganizationLayout from "./NoOrganizationLayout";
import OrganizationLayout from "./OrganizationLayout";
import { SettingsPage } from "./SettingsPage";
import { EditFrameworkPage } from "./frameworks/EditFrameworkPage";
import { FrameworkLayout } from "./frameworks/FrameworkLayout";
import { FrameworkListPage } from "./frameworks/FrameworkListPage";
import { FrameworkPage } from "./frameworks/FrameworkPage";
import { NewFrameworkPage } from "./frameworks/NewFrameworkPage";
import { ControlPage } from "./frameworks/controls/ControlPage";
import { EditMeasurePage } from "./measures/EditMeasurePage";
import { MeasureListPage } from "./measures/MeasureListPage";
import { MeasurePage } from "./measures/MeasurePage";
import { NewMeasurePage } from "./measures/NewMeasurePage";
import { NewPeoplePage } from "./people/NewPeoplePage";
import { PeopleListPage } from "./people/PeopleListPage";
import { PeoplePage } from "./people/PeoplePage";
import { EditDocumentPage } from "./documents/EditDocumentPage";
import { DocumentListPage } from "./documents/DocumentListPage";
import { ShowDocumentPage } from "./documents/ShowDocumentPage";
import { EditRiskPage } from "./risks/EditRiskPage";
import { NewRiskPage } from "./risks/NewRiskPage";
import { ListRiskPage } from "./risks/ListRiskPage";
import ShowRiskView from "./risks/ShowRiskView";
import { ListVendorPage } from "./vendors/ListVendorPage";
import { VendorPage } from "./vendors/VendorPage";
import { ListTaskPage } from "./tasks/ListTaskPage";
import { AssetsListPage } from "./assets/AssetsListPage";
import { NewAssetPage } from "./assets/NewAssetPage";
import { AssetPage } from "./assets/AssetPage";
import { DataListPage } from "./data/DataListPage";
import { NewDatumPage } from "./data/NewDatumPage";
import { DatumPage } from "./data/DatumPage";
import { EditControlPage } from "./frameworks/controls/EditControlPage";
import { NewControlPage } from "./frameworks/controls/NewControlPage";

export function OrganizationsRoutes() {
  return (
    <Routes>
      <Route path=":organizationId/*" element={<OrganizationLayout />}>
        <Route index element={<HomePage />} />
        <Route path="tasks" element={<ListTaskPage />} />
        <Route path="people" element={<PeopleListPage />} />
        <Route path="people/new" element={<NewPeoplePage />} />
        <Route path="people/:peopleId" element={<PeoplePage />} />
        <Route path="vendors" element={<ListVendorPage />} />
        <Route path="frameworks" element={<FrameworkListPage />} />
        <Route path="frameworks/new" element={<NewFrameworkPage />} />
        <Route path="frameworks/:frameworkId/*">
          <Route element={<FrameworkLayout />}>
            <Route index element={<FrameworkPage />} />
            <Route path="controls/new" element={<NewControlPage />} />
            <Route path="controls/:controlId" element={<ControlPage />} />
            <Route path="controls/:controlId/edit" element={<EditControlPage />} />
          </Route>
          <Route path="edit" element={<EditFrameworkPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Route>
        <Route path="measures" element={<MeasureListPage />} />
        <Route path="measures/new" element={<NewMeasurePage />} />
        <Route path="measures/:measureId" element={<MeasurePage />} />
        <Route path="measures/:measureId/edit" element={<EditMeasurePage />} />
        <Route path="vendors/:vendorId" element={<VendorPage />} />
        <Route path="documents" element={<DocumentListPage />} />
        <Route path="documents/:documentId" element={<ShowDocumentPage />} />
        <Route
          path="documents/:documentId/versions/:versionId/edit"
          element={<EditDocumentPage />}
        />
        <Route path="risks" element={<ListRiskPage />} />
        <Route path="risks/new" element={<NewRiskPage />} />
        <Route path="risks/:riskId" element={<ShowRiskView />} />
        <Route path="risks/:riskId/edit" element={<EditRiskPage />} />
        <Route path="assets" element={<AssetsListPage />} />
        <Route path="assets/new" element={<NewAssetPage />} />
        <Route path="assets/:assetId" element={<AssetPage />} />
        <Route path="data" element={<DataListPage />} />
        <Route path="data/new" element={<NewDatumPage />} />
        <Route path="data/:datumId" element={<DatumPage />} />
        <Route path="settings" element={<SettingsPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>

      <Route path="*" element={<NoOrganizationLayout />}>
        <Route path="new" element={<CreateOrganizationPage />} />
      </Route>
    </Routes>
  );
}
