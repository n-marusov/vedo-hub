// @ctx: M2.5 E2E test JWT tokens — pre-signed RS256 JWTs for real-backend auth.
//
// These JWTs are signed with the test RSA private key in test-jwt-key.pem.
// The corresponding public key is set as JWT_DEV_PUBLIC_KEY_PEM in
// deploy/docker-compose.test.yml for the api-gateway-test service.
//
// Regenerate tokens when the keypair or claims change:
//   node tests/e2e/playwright/sign-jwt.js

// Owner token with Owner role for full CRUD access
export const OWNER_JWT =
  'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyIsInVzZXJfaWQiOiJ1c2VyLTEyMyIsInRlbmFudF9pZCI6IiIsInJvbGVzIjpbIk93bmVyIl0sImlhdCI6MTc4NDQwNzEyNiwiZXhwIjoxNzg0NDkzNTI2fQ.LmRa3PJIjVCsWYksprNbvMLDyTwzL7HG-Y5DPVXGf_gMLf8xquNOEuTaK7JH-bEoWiLOMlll5EkuLSqtfHSKJY9XdD6J6GSoAcE4zckp2N_PZ-rfVIrXrClHVi-e301DTWeYHy2-GOKRt5ovO7MJ55UgiczFqa4HBIqVS7KVSzQ5gWkUzQUPP3wpFpbQuFdDBtPy-St0tpBZ3QBba9ISGNkzZx-masgXWh_6kSOGm9aioRslaca60MtayYbNlI-pk_bBSZmxxdmRv78XCpKU0EAWagFqPpp8rMZb4-DOw0UXxE_ECUXX8aedh7ENyG0yX6tFFZqF0-PkfoeGoFwoxw';

// Viewer token with Viewer role for 403 authorization tests
export const VIEWER_JWT =
  'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ2aWV3ZXItdXNlciIsInVzZXJfaWQiOiJ2aWV3ZXItdXNlciIsInRlbmFudF9pZCI6IiIsInJvbGVzIjpbIlZpZXdlciJdLCJpYXQiOjE3ODQ0MDcxMjYsImV4cCI6MTc4NDQ5MzUyNn0.ToHcY3dQweRcQZSqP7BgkjUH263BkqFklNWrQ9o-ut-Wa7J-I4Lq8mLYWBOKCVsI05YywimDz0JmVlpFuKi4WtvOsGsB7Y4LPkcciRcPz17y6HQGG3GrvmhZ3oenMceDtDSsbJakBCxzVEec_AwEiUBCMej6E7fvTreNPlU3C9jg_Ebp8vWw0AhqgNZl6E7kYuZo1_2DQnA6Mcx9ZROz_Dn9P7YdUYbVaF99GyI8pP39mM4TOfq7meT4cjPchYUsLTP6_pATQkDjmnTCxc15_86thcO3NOs2yVnHK1993ewNn_cnsSW8sFKe7-Ml-PsSUwMM_OqQynwfQ3wZXZIvWA';
