import {
  Navigate,
  Route,
  RouterProvider,
  createBrowserRouter,
  createRoutesFromElements,
} from "react-router-dom";
import { Provider as GraphqlProvider } from "urql";

import { graphqlClient } from "./graphql";
import { AppLayout } from "./layouts/AppLayout/AppLayout";
import { ChatPage } from "./pages/ChatPage/ChatPage";
import { LoginPage } from "./pages/LoginPage/LoginPage";
import { ProtectedRoute } from "./components/ProtectedRoute/ProtectedRoute";
import "./styles/font.css.ts";
import "./styles/global.css.ts";
import "./styles/theme.css.ts";

const router = createBrowserRouter(
  createRoutesFromElements(
    <>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<AppLayout />}>
          <Route path="/chat/:id?" element={<ChatPage />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/chat" />} />
    </>,
  ),
);

function App() {
  return (
    <GraphqlProvider value={graphqlClient}>
      <RouterProvider router={router} />
    </GraphqlProvider>
  );
}

export default App;
