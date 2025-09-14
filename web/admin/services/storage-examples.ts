/**
 * StorageClient Usage Examples
 * 
 * This file demonstrates how to use the StorageClient class to interact
 * with the Goro storage API for various use cases.
 */

import StorageClient, { 
  StorageResource, 
  StorageResourceResponse,
  ContentNegotiationOptions,
  UploadOptions 
} from './resource';

// Initialize the client
const storageClient = new StorageClient({
  baseURL: 'http://localhost:8080', // Adjust to your server URL
  timeout: 30000, // 30 seconds
  defaultHeaders: {
    'User-Agent': 'Goro-Admin-Client/1.0'
  }
});

/**
 * Example 1: Basic CRUD Operations
 */
export async function basicCrudExample() {
  try {
    // Create a simple text resource
    const textData = 'Hello, World! This is a test resource.';
    const createResponse = await storageClient.createResource(
      textData,
      { contentType: 'text/plain' }
    );
    console.log('Created resource:', createResponse);

    // Get the resource back
    const resource = await storageClient.getResource(createResponse.id);
    console.log('Retrieved resource:', {
      id: resource.id,
      contentType: resource.contentType,
      size: resource.size,
      data: new TextDecoder().decode(resource.data as Uint8Array)
    });

    // Update the resource
    const updatedData = 'Updated content for the resource.';
    const updateResponse = await storageClient.putResource(
      createResponse.id,
      updatedData,
      { contentType: 'text/plain' }
    );
    console.log('Updated resource:', updateResponse);

    // Check resource metadata
    const metadata = await storageClient.getResourceMetadata(createResponse.id);
    console.log('Resource metadata:', metadata);

    // Delete the resource
    const deleteResponse = await storageClient.deleteResource(createResponse.id);
    console.log('Deleted resource:', deleteResponse);

  } catch (error) {
    console.error('CRUD operation failed:', error);
  }
}

/**
 * Example 2: RDF Content with Content Negotiation
 */
export async function rdfContentExample() {
  try {
    // Create JSON-LD resource
    const jsonLdData = {
      "@context": "https://schema.org/",
      "@type": "Person",
      "name": "John Doe",
      "email": "john@example.com"
    };

    const jsonLdResponse = await storageClient.createResource(
      JSON.stringify(jsonLdData, null, 2),
      { contentType: 'application/ld+json' }
    );
    console.log('Created JSON-LD resource:', jsonLdResponse);

    // Retrieve as different RDF formats using content negotiation
    const asJsonLd = await storageClient.getResource(
      jsonLdResponse.id,
      { preferredFormat: 'application/ld+json' }
    );
    console.log('As JSON-LD:', new TextDecoder().decode(asJsonLd.data as Uint8Array));

    const asTurtle = await storageClient.getResource(
      jsonLdResponse.id,
      { preferredFormat: 'text/turtle' }
    );
    console.log('As Turtle:', new TextDecoder().decode(asTurtle.data as Uint8Array));

    const asRdfXml = await storageClient.getResource(
      jsonLdResponse.id,
      { preferredFormat: 'application/rdf+xml' }
    );
    console.log('As RDF/XML:', new TextDecoder().decode(asRdfXml.data as Uint8Array));

  } catch (error) {
    console.error('RDF operation failed:', error);
  }
}

/**
 * Example 3: Large File Upload with Streaming
 */
export async function largeFileUploadExample() {
  try {
    // Simulate a large file (in real usage, this would come from a file input)
    const largeData = new Uint8Array(2 * 1024 * 1024); // 2MB of data
    for (let i = 0; i < largeData.length; i++) {
      largeData[i] = i % 256;
    }

    // Upload using streaming for large files
    const resourceId = 'large-file-example';
    const uploadResponse = await storageClient.putResource(
      resourceId,
      largeData,
      { 
        contentType: 'application/octet-stream',
        useStreaming: true,
        contentLength: largeData.length
      }
    );
    console.log('Uploaded large file:', uploadResponse);

    // Stream the resource back
    const stream = await storageClient.streamResource(resourceId);
    
    // Process the stream (example: calculate size)
    const reader = stream.getReader();
    let totalSize = 0;
    
    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        totalSize += value.length;
      }
      console.log(`Streamed ${totalSize} bytes`);
    } finally {
      reader.releaseLock();
    }

  } catch (error) {
    console.error('Large file operation failed:', error);
  }
}

/**
 * Example 4: File Upload from Browser
 */
export async function fileUploadExample(file: File) {
  try {
    // Check if the content type is supported
    const normalizedType = storageClient.normalizeContentType(file.type);
    
    if (storageClient.isFormatSupported(normalizedType)) {
      console.log(`Uploading ${file.name} as supported RDF format: ${normalizedType}`);
    } else {
      console.log(`Uploading ${file.name} as binary data: ${file.type}`);
    }

    // Generate a unique resource ID
    const resourceId = `upload-${Date.now()}-${file.name}`;

    // Upload the file
    const uploadResponse = await storageClient.putResource(
      resourceId,
      file,
      { 
        contentType: file.type || 'application/octet-stream',
        useStreaming: file.size > 1024 * 1024, // Use streaming for files > 1MB
        contentLength: file.size
      }
    );

    console.log('File uploaded successfully:', uploadResponse);
    return uploadResponse;

  } catch (error) {
    console.error('File upload failed:', error);
    throw error;
  }
}

/**
 * Example 5: Resource Management
 */
export async function resourceManagementExample() {
  try {
    // Check what operations are supported
    const options = await storageClient.getResourceOptions();
    console.log('Supported operations:', options);

    // Create multiple resources
    const resourceIds: string[] = [];
    
    for (let i = 1; i <= 3; i++) {
      const data = `Test resource ${i}`;
      const response = await storageClient.createResource(
        data,
        { contentType: 'text/plain' },
        `test-resource-${i}`
      );
      resourceIds.push(response.id);
      console.log(`Created resource ${i}:`, response.id);
    }

    // Check which resources exist
    for (const id of resourceIds) {
      const exists = await storageClient.resourceExists(id);
      console.log(`Resource ${id} exists:`, exists);
      
      if (exists) {
        const metadata = await storageClient.getResourceMetadata(id);
        console.log(`Resource ${id} metadata:`, metadata);
      }
    }

    // Clean up - delete all test resources
    for (const id of resourceIds) {
      try {
        await storageClient.deleteResource(id);
        console.log(`Deleted resource: ${id}`);
      } catch (error) {
        console.error(`Failed to delete resource ${id}:`, error);
      }
    }

  } catch (error) {
    console.error('Resource management failed:', error);
  }
}

/**
 * Example 6: Error Handling
 */
export async function errorHandlingExample() {
  try {
    // Try to get a non-existent resource
    await storageClient.getResource('non-existent-resource');
  } catch (error: any) {
    console.log('Expected error for non-existent resource:', error.message);
    
    // Check if it's a storage error with additional details
    if (error.storageError) {
      console.log('Storage error details:', {
        code: error.storageError.code,
        operation: error.storageError.operation,
        context: error.storageError.context
      });
    }
  }

  try {
    // Try to upload with an unsupported format
    await storageClient.createResource(
      'Invalid data',
      { contentType: 'application/unsupported-format' }
    );
  } catch (error: any) {
    console.log('Expected error for unsupported format:', error.message);
    
    if (error.storageError?.supportedFormats) {
      console.log('Supported formats:', error.storageError.supportedFormats);
    }
  }
}

/**
 * Example 7: Batch Operations
 */
export async function batchOperationsExample() {
  const resources = [
    { id: 'batch-1', data: 'First resource', type: 'text/plain' },
    { id: 'batch-2', data: 'Second resource', type: 'text/plain' },
    { id: 'batch-3', data: 'Third resource', type: 'text/plain' }
  ];

  try {
    // Upload multiple resources in parallel
    const uploadPromises = resources.map(resource =>
      storageClient.putResource(
        resource.id,
        resource.data,
        { contentType: resource.type }
      )
    );

    const uploadResults = await Promise.allSettled(uploadPromises);
    
    uploadResults.forEach((result, index) => {
      if (result.status === 'fulfilled') {
        console.log(`Successfully uploaded ${resources[index].id}`);
      } else {
        console.error(`Failed to upload ${resources[index].id}:`, result.reason);
      }
    });

    // Check all resources exist
    const existsPromises = resources.map(resource =>
      storageClient.resourceExists(resource.id)
    );

    const existsResults = await Promise.all(existsPromises);
    console.log('Resources exist:', existsResults);

    // Clean up
    const deletePromises = resources.map(resource =>
      storageClient.deleteResource(resource.id).catch(err => console.error(`Delete failed for ${resource.id}:`, err))
    );

    await Promise.all(deletePromises);
    console.log('Cleanup completed');

  } catch (error) {
    console.error('Batch operations failed:', error);
  }
}

// Utility function to format bytes
export function formatBytes(bytes: number, decimals: number = 2): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

// Export the configured client for use in other modules
export { storageClient };
export default storageClient;
