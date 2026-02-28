/**
 * Test file for Secure Storage
 * Run these tests manually in the browser console
 */

import { secureStorage, isTokenExpired, decodeToken, validateTokenFormat } from './secureStorage';

/**
 * Manual test suite for secure storage
 * Copy and paste into browser console to test
 */
export async function testSecureStorage() {
  console.log('=== Secure Storage Tests ===\n');

  // Test 1: Basic encryption/decryption
  console.log('Test 1: Basic encryption/decryption');
  try {
    await secureStorage.setItem('test_key', 'Hello, World!');
    const value = await secureStorage.getItem('test_key');
    console.assert(value === 'Hello, World!', 'Value should match');
    console.log('✓ Basic encryption/decryption works\n');
  } catch (error) {
    console.error('✗ Basic encryption/decryption failed:', error);
  }

  // Test 2: Non-existent key
  console.log('Test 2: Non-existent key');
  try {
    const value = await secureStorage.getItem('non_existent_key');
    console.assert(value === null, 'Should return null for non-existent key');
    console.log('✓ Non-existent key returns null\n');
  } catch (error) {
    console.error('✗ Non-existent key test failed:', error);
  }

  // Test 3: Token storage
  console.log('Test 3: Token storage');
  try {
    const fakeToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c';
    await secureStorage.setItem('access_token', fakeToken);
    const retrieved = await secureStorage.getItem('access_token');
    console.assert(retrieved === fakeToken, 'Token should match');
    console.log('✓ Token storage works\n');
  } catch (error) {
    console.error('✗ Token storage failed:', error);
  }

  // Test 4: Verify data is encrypted in sessionStorage
  console.log('Test 4: Verify data is encrypted');
  try {
    await secureStorage.setItem('plain_test', 'SensitiveData123');
    const rawValue = sessionStorage.getItem('__secure__plain_test');
    console.log('Raw encrypted value:', rawValue);
    console.assert(rawValue !== 'SensitiveData123', 'Data should be encrypted');
    console.assert(rawValue !== btoa('SensitiveData123'), 'Data should not be simple base64');
    console.log('✓ Data is properly encrypted\n');
  } catch (error) {
    console.error('✗ Encryption verification failed:', error);
  }

  // Test 5: Clear storage
  console.log('Test 5: Clear storage');
  try {
    await secureStorage.setItem('clear_test', 'will be deleted');
    secureStorage.clear();
    const value = await secureStorage.getItem('clear_test');
    console.assert(value === null, 'Value should be cleared');
    console.log('✓ Clear storage works\n');
  } catch (error) {
    console.error('✗ Clear storage failed:', error);
  }

  // Test 6: Unique encryption (same data, different ciphertext)
  console.log('Test 6: Unique encryption with random IV');
  try {
    const testData = 'TestData123';
    await secureStorage.setItem('unique_test_1', testData);
    const encrypted1 = sessionStorage.getItem('__secure__unique_test_1');

    secureStorage.removeItem('unique_test_1');
    await secureStorage.setItem('unique_test_1', testData);
    const encrypted2 = sessionStorage.getItem('__secure__unique_test_1');

    console.assert(encrypted1 !== encrypted2, 'Same data should have different ciphertext due to random IV');
    console.log('✓ Unique encryption works (random IV)\n');
  } catch (error) {
    console.error('✗ Unique encryption test failed:', error);
  }

  // Test 7: Token validation
  console.log('Test 7: Token validation utilities');
  const validToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c';
  console.assert(validateTokenFormat(validToken) === true, 'Valid JWT format should pass');
  console.assert(validateTokenFormat('invalid') === false, 'Invalid format should fail');
  console.assert(isTokenExpired(validToken) === true, 'Old token should be expired');
  console.log('✓ Token validation works\n');

  console.log('=== All Tests Complete ===');
}

// Browser console helper
if (typeof window !== 'undefined') {
  (window as any).testSecureStorage = testSecureStorage;
  console.log('Run testSecureStorage() in console to test');
}
