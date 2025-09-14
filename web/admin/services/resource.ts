// TypeScript interfaces for Storage API
export interface StorageResource {
  id: string;
  contentType: string;
  size: number;
  data?: Uint8Array | string;
  metadata?: Record<string, any>;
}

export interface StorageResourceResponse {
  id: string;
  contentType: string;
  size: number;
  message: string;
  streaming?: boolean;
}

export interface StorageError {
  code: string;
  message: string;
  status: number;
  timestamp: string;
  operation?: string;
  context?: Record<string, any>;
  details?: string;
  supportedFormats?: string[];
  suggestion?: string;
}

export interface StorageErrorResponse {
  error: StorageError;
}

export interface StreamMetadata {
  contentType: string;
  size?: number;
}

export interface ContentNegotiationOptions {
  accept?: string;
  preferredFormat?: 'application/ld+json' | 'text/turtle' | 'application/rdf+xml';
}

export interface ContainerInfo {
  id: string;
  title?: string;
  description?: string;
  memberCount: number;
  members: string[];
  pagination?: {
    limit: number;
    offset: number;
  };
}

export interface TreeNode {
  key: string;
  title: string;
  type: 'container' | 'resource';
  size?: number;
  contentType?: string;
  children?: TreeNode[];
  isLeaf?: boolean;
}

export interface UploadOptions {
  contentType: string;
  useStreaming?: boolean;
  contentLength?: number;
}

export interface StorageClientConfig {
  baseURL: string;
  timeout?: number;
  defaultHeaders?: Record<string, string>;
}

/**
 * Storage API Client for interacting with the Goro storage service
 * 
 * Provides a comprehensive interface for:
 * - CRUD operations on resources
 * - Streaming support for large files
 * - Content negotiation for RDF formats
 * - Comprehensive error handling
 * - Type-safe API interactions
 */
export class StorageClient {
  private baseURL: string;
  private timeout: number;
  private defaultHeaders: Record<string, string>;

  // Supported RDF formats for content negotiation
  private static readonly SUPPORTED_FORMATS = [
    'application/ld+json',
    'text/turtle', 
    'application/rdf+xml'
  ] as const;

  constructor(config: StorageClientConfig) {
    this.baseURL = config.baseURL.replace(/\/$/, ''); // Remove trailing slash
    this.timeout = config.timeout || 30000; // 30 second default timeout
    this.defaultHeaders = {
      'Content-Type': 'application/json',
      ...config.defaultHeaders
    };
  }

  /**
   * Get a resource by ID with optional content negotiation
   */
  async getResource(
    id: string, 
    options: ContentNegotiationOptions = {}
  ): Promise<StorageResource> {
    const headers: Record<string, string> = { ...this.defaultHeaders };
    
    if (options.accept) {
      headers['Accept'] = options.accept;
    } else if (options.preferredFormat) {
      headers['Accept'] = options.preferredFormat;
    }

    const response = await this.makeRequest(`/resources/${encodeURIComponent(id)}`, {
      method: 'GET',
      headers
    });

    if (!response.ok) {
      throw await this.handleErrorResponse(response);
    }

    const data = await response.arrayBuffer();
    const contentType = response.headers.get('Content-Type') || 'application/octet-stream';
    const contentLength = response.headers.get('Content-Length');

    return {
      id,
      contentType,
      size: contentLength ? parseInt(contentLength) : data.byteLength,
      data: new Uint8Array(data)
    };
  }

  /**
   * Create a new resource (POST to collection)
   */
  async createResource(
    data: Uint8Array | string | Blob,
    options: UploadOptions,
    resourceId?: string
  ): Promise<StorageResourceResponse> {
    const url = resourceId ? `/resources/${encodeURIComponent(resourceId)}` : '/resources';
    
    const headers: Record<string, string> = {
      ...this.defaultHeaders,
      'Content-Type': options.contentType
    };

    // Handle different data types
    let body: BodyInit;
    if (data instanceof Uint8Array) {
      body = data;
      headers['Content-Length'] = data.byteLength.toString();
    } else if (typeof data === 'string') {
      body = data;
      headers['Content-Length'] = new TextEncoder().encode(data).byteLength.toString();
    } else if (data instanceof Blob) {
      body = data;
      headers['Content-Length'] = data.size.toString();
    } else {
      throw new Error('Unsupported data type');
    }

    // Use streaming for large files if requested or if size exceeds threshold
    const shouldUseStreaming = options.useStreaming || 
      (options.contentLength && options.contentLength > 1024 * 1024);

    const response = await this.makeRequest(url, {
      method: 'POST',
      headers,
      body
    });

    if (!response.ok) {
      throw await this.handleErrorResponse(response);
    }

    return await response.json();
  }

  /**
   * Update or create a resource (PUT)
   */
  async putResource(
    id: string,
    data: Uint8Array | string | Blob,
    options: UploadOptions
  ): Promise<StorageResourceResponse> {
    const headers: Record<string, string> = {
      ...this.defaultHeaders,
      'Content-Type': options.contentType
    };

    let body: BodyInit;
    if (data instanceof Uint8Array) {
      body = data;
      headers['Content-Length'] = data.byteLength.toString();
    } else if (typeof data === 'string') {
      body = data;
      headers['Content-Length'] = new TextEncoder().encode(data).byteLength.toString();
    } else if (data instanceof Blob) {
      body = data;
      headers['Content-Length'] = data.size.toString();
    } else {
      throw new Error('Unsupported data type');
    }

    const response = await this.makeRequest(`/resources/${encodeURIComponent(id)}`, {
      method: 'PUT',
      headers,
      body
    });

    if (!response.ok) {
      throw await this.handleErrorResponse(response);
    }

    return await response.json();
  }

  /**
   * Delete a resource
   */
  async deleteResource(id: string): Promise<{ id: string; message: string }> {
    const response = await this.makeRequest(`/resources/${encodeURIComponent(id)}`, {
      method: 'DELETE',
      headers: this.defaultHeaders
    });

    if (!response.ok) {
      throw await this.handleErrorResponse(response);
    }

    return await response.json();
  }

  /**
   * Get resource metadata (HEAD request)
   */
  async getResourceMetadata(
    id: string,
    options: ContentNegotiationOptions = {}
  ): Promise<StreamMetadata> {
    const headers: Record<string, string> = { ...this.defaultHeaders };
    
    if (options.accept) {
      headers['Accept'] = options.accept;
    } else if (options.preferredFormat) {
      headers['Accept'] = options.preferredFormat;
    }

    const response = await this.makeRequest(`/resources/${encodeURIComponent(id)}`, {
      method: 'HEAD',
      headers
    });

    if (!response.ok) {
      throw await this.handleErrorResponse(response);
    }

    const contentType = response.headers.get('Content-Type') || 'application/octet-stream';
    const contentLength = response.headers.get('Content-Length');

    return {
      contentType,
      size: contentLength ? parseInt(contentLength) : undefined
    };
  }

  /**
   * Get supported methods and formats for resources (OPTIONS request)
   */
  async getResourceOptions(id?: string): Promise<{
    methods: string[];
    formats: string[];
  }> {
    const url = id ? `/resources/${encodeURIComponent(id)}` : '/resources';
    
    const response = await this.makeRequest(url, {
      method: 'OPTIONS',
      headers: this.defaultHeaders
    });

    if (!response.ok) {
      throw await this.handleErrorResponse(response);
    }

    return await response.json();
  }

  /**
   * Stream a resource for large files
   */
  async streamResource(
    id: string,
    options: ContentNegotiationOptions = {}
  ): Promise<ReadableStream<Uint8Array>> {
    const headers: Record<string, string> = { ...this.defaultHeaders };
    
    if (options.accept) {
      headers['Accept'] = options.accept;
    } else if (options.preferredFormat) {
      headers['Accept'] = options.preferredFormat;
    }

    const response = await this.makeRequest(`/resources/${encodeURIComponent(id)}`, {
      method: 'GET',
      headers
    });

    if (!response.ok) {
      throw await this.handleErrorResponse(response);
    }

    if (!response.body) {
      throw new Error('Response body is not available for streaming');
    }

    return response.body;
  }

  /**
   * Upload a large file using streaming
   */
  async uploadResourceStream(
    id: string,
    stream: ReadableStream<Uint8Array>,
    options: UploadOptions
  ): Promise<StorageResourceResponse> {
    const headers: Record<string, string> = {
      ...this.defaultHeaders,
      'Content-Type': options.contentType,
      'Transfer-Encoding': 'chunked'
    };

    if (options.contentLength) {
      headers['Content-Length'] = options.contentLength.toString();
    }

    const response = await this.makeRequest(`/resources/${encodeURIComponent(id)}`, {
      method: 'PUT',
      headers,
      body: stream
    });

    if (!response.ok) {
      throw await this.handleErrorResponse(response);
    }

    return await response.json();
  }

  /**
   * Check if a resource exists
   */
  async resourceExists(id: string): Promise<boolean> {
    try {
      await this.getResourceMetadata(id);
      return true;
    } catch (error) {
      if (error instanceof Error && error.message.includes('404')) {
        return false;
      }
      throw error;
    }
  }

  /**
   * Validate if a format is supported
   */
  isFormatSupported(format: string): boolean {
    return StorageClient.SUPPORTED_FORMATS.includes(format as any);
  }

  /**
   * Get list of supported formats
   */
  getSupportedFormats(): readonly string[] {
    return StorageClient.SUPPORTED_FORMATS;
  }

  /**
   * Get container information including its members
   */
  async getContainer(
    id: string, 
    options: ContentNegotiationOptions = {}
  ): Promise<ContainerInfo> {
    const headers: Record<string, string> = { ...this.defaultHeaders };
    
    if (options.accept) {
      headers['Accept'] = options.accept;
    } else if (options.preferredFormat) {
      headers['Accept'] = options.preferredFormat;
    }

    const response = await this.makeRequest(`/containers/${encodeURIComponent(id)}`, {
      method: 'GET',
      headers
    });

    if (!response.ok) {
      throw await this.handleErrorResponse(response);
    }

    const data = await response.json();
    
    return {
      id: data['@id'] || id,
      title: data['dcterms:title'] || data.title,
      description: data['dcterms:description'] || data.description,
      memberCount: data['ldp:memberCount'] || data.memberCount || 0,
      members: data['ldp:contains'] || data.members || [],
      pagination: data.pagination
    };
  }

  /**
   * Get children of a container (sub-containers)
   */
  async getContainerChildren(id: string): Promise<ContainerInfo[]> {
    try {
      const response = await this.makeRequest(`/containers/${encodeURIComponent(id)}/children`, {
        method: 'GET',
        headers: this.defaultHeaders
      });

      if (!response.ok) {
        // If endpoint doesn't exist, return empty array
        if (response.status === 404) {
          return [];
        }
        throw await this.handleErrorResponse(response);
      }

      const data = await response.json();
      return data.children || [];
    } catch (error) {
      // Fallback to empty array if children endpoint is not implemented
      console.warn('Container children endpoint not available:', error);
      return [];
    }
  }

  /**
   * Scan all known containers and resources to build a tree structure
   * This is a fallback method when hierarchical APIs aren't available
   */
  async scanAllResources(): Promise<TreeNode[]> {
    // First, let's try to find some containers by scanning common IDs
    const knownContainerIds = ['root', 'main', 'data', 'files', 'documents', 'images'];
    const foundContainers: ContainerInfo[] = [];
    const foundResources: StorageResource[] = [];

    // Try to find containers
    for (const id of knownContainerIds) {
      try {
        const container = await this.getContainer(id);
        foundContainers.push(container);
      } catch (error) {
        // Container doesn't exist, continue
      }
    }

    // If no containers found, try to find some resources
    if (foundContainers.length === 0) {
      const resourceIds = ['test', 'example', 'sample', 'demo'];
      for (const id of resourceIds) {
        try {
          const resource = await this.getResource(id);
          foundResources.push(resource);
        } catch (error) {
          // Resource doesn't exist, continue
        }
      }
    }

    // Build tree structure
    return this.buildTree(foundContainers, foundResources);
  }

  /**
   * Build a tree structure from containers and resources
   */
  private async buildTree(containers: ContainerInfo[], resources: StorageResource[]): Promise<TreeNode[]> {
    const treeNodes: TreeNode[] = [];

    // Add containers
    for (const container of containers) {
      const containerNode: TreeNode = {
        key: container.id,
        title: container.title || container.id,
        type: 'container',
        children: [],
        isLeaf: false
      };

      // Add container members as children
      for (const memberId of container.members) {
        try {
          // Try to get the member as a resource first
          const resource = await this.getResource(memberId);
          containerNode.children!.push({
            key: memberId,
            title: memberId,
            type: 'resource',
            size: resource.size,
            contentType: resource.contentType,
            isLeaf: true
          });
        } catch (error) {
          // If it's not a resource, it might be another container
          try {
            const subContainer = await this.getContainer(memberId);
            const subContainerNode: TreeNode = {
              key: memberId,
              title: subContainer.title || memberId,
              type: 'container',
              children: [],
              isLeaf: false
            };
            containerNode.children!.push(subContainerNode);
          } catch (subError) {
            // If we can't fetch it, just add it as a resource placeholder
            containerNode.children!.push({
              key: memberId,
              title: memberId,
              type: 'resource',
              isLeaf: true
            });
          }
        }
      }

      treeNodes.push(containerNode);
    }

    // Add standalone resources (not in any container)
    for (const resource of resources) {
      treeNodes.push({
        key: resource.id,
        title: resource.id,
        type: 'resource',
        size: resource.size,
        contentType: resource.contentType,
        isLeaf: true
      });
    }

    return treeNodes;
  }

  /**
   * Get a full tree view of all containers and their contents
   */
  async getFullTree(): Promise<TreeNode[]> {
    try {
      // Try to get a root container first
      const rootContainer = await this.getContainer('root');
      return await this.buildTreeFromContainer(rootContainer);
    } catch (error) {
      // If no root container, scan for available resources
      console.warn('No root container found, scanning for available resources...');
      return await this.scanAllResources();
    }
  }

  /**
   * Build tree recursively from a container
   */
  private async buildTreeFromContainer(container: ContainerInfo, visited: Set<string> = new Set()): Promise<TreeNode[]> {
    // Prevent infinite recursion
    if (visited.has(container.id)) {
      return [];
    }
    visited.add(container.id);

    const node: TreeNode = {
      key: container.id,
      title: container.title || container.id,
      type: 'container',
      children: [],
      isLeaf: false
    };

    // Process each member
    for (const memberId of container.members) {
      try {
        // Try as container first
        const memberContainer = await this.getContainer(memberId);
        const childNodes = await this.buildTreeFromContainer(memberContainer, visited);
        node.children!.push(...childNodes);
      } catch (error) {
        // Try as resource
        try {
          const resource = await this.getResource(memberId);
          node.children!.push({
            key: memberId,
            title: memberId,
            type: 'resource',
            size: resource.size,
            contentType: resource.contentType,
            isLeaf: true
          });
        } catch (resourceError) {
          // Add as unknown item
          node.children!.push({
            key: memberId,
            title: memberId,
            type: 'resource',
            isLeaf: true
          });
        }
      }
    }

    return [node];
  }

  /**
   * Convert content type aliases to standard formats
   */
  normalizeContentType(contentType: string): string {
    const normalized = contentType.toLowerCase().trim();
    
    switch (normalized) {
      case 'json-ld':
      case 'jsonld':
      case 'application/json':
        return 'application/ld+json';
      case 'turtle':
      case 'ttl':
      case 'text/plain':
        return 'text/turtle';
      case 'rdf/xml':
      case 'rdfxml':
      case 'xml':
        return 'application/rdf+xml';
      default:
        return normalized;
    }
  }

  /**
   * Make HTTP request with error handling and timeouts
   */
  private async makeRequest(path: string, options: RequestInit): Promise<Response> {
    const url = `${this.baseURL}${path}`;
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const response = await fetch(url, {
        ...options,
        signal: controller.signal
      });
      clearTimeout(timeoutId);
      return response;
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof Error && error.name === 'AbortError') {
        throw new Error(`Request timeout after ${this.timeout}ms`);
      }
      throw error;
    }
  }

  /**
   * Handle error responses and convert them to typed errors
   */
  private async handleErrorResponse(response: Response): Promise<Error> {
    try {
      const errorData: StorageErrorResponse = await response.json();
      const error = errorData.error;
      
      let message = `${error.code}: ${error.message}`;
      if (error.details) {
        message += ` - ${error.details}`;
      }
      if (error.suggestion) {
        message += ` (${error.suggestion})`;
      }

      const customError = new Error(message);
      (customError as any).storageError = error;
      (customError as any).status = response.status;
      
      return customError;
    } catch (parseError) {
      // If we can't parse the error response, return a generic error
      return new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
  }
}

// Default export for convenience
export default StorageClient;
