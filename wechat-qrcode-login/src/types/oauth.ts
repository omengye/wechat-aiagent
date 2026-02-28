/**
 * OAuth Authorization Types
 */

export interface OAuthAuthorizeParams {
  client_id: string;
  redirect_url: string;
  response_type: 'code';
  scope: string;
  state?: string;
  code_challenge?: string;
  code_challenge_method?: 'S256' | 'plain';
}

export interface OAuthAuthorizeResponse {
  session_id: string;
  client_name: string;
  scope_names: string[];
  scopes: string;
  project_id: string;
  redirect_url: string;
  state: string;
}

export interface OAuthError {
  error: string;
  error_description: string;
}
