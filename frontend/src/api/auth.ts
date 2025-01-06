export const logout = async (returnUri: string) => {
  const params = new URLSearchParams({ return_uri: returnUri });
  window.location.href = `/api/v1/auth/logout?${params}`;
}; 