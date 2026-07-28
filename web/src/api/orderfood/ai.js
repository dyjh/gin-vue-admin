import { orderFoodRequest } from './request'

const encodePath = (value) => encodeURIComponent(value)

// 平台整体能力策略
export const getPlatformCapabilityPolicy = () =>
  orderFoodRequest({
    path: '/platform-capability-policy',
    method: 'get'
  })

export const updatePlatformCapabilityPolicy = (data) =>
  orderFoodRequest({
    path: '/platform-capability-policy',
    method: 'put',
    data,
    mutation: true
  })

export const updatePlatformCapabilityEmergencyStatus = (data) =>
  orderFoodRequest({
    path: '/platform-capability-policy/emergency-status',
    method: 'put',
    data,
    mutation: true
  })

// AI 供应商
export const getAIProviderList = (params) =>
  orderFoodRequest({
    path: '/ai-providers',
    method: 'get',
    params
  })

export const getAIProviderDetail = (providerId) =>
  orderFoodRequest({
    path: `/ai-providers/${encodePath(providerId)}`,
    method: 'get'
  })

export const getAIProviderModelOptions = (providerId) =>
  orderFoodRequest({
    path: `/ai-providers/${encodePath(providerId)}/models`,
    method: 'get'
  })

export const getAIModelProviderOptions = () =>
  orderFoodRequest({
    path: '/ai-model-provider-options',
    method: 'get'
  })

export const createAIProvider = (data) =>
  orderFoodRequest({
    path: '/ai-providers',
    method: 'post',
    data,
    mutation: true
  })

export const updateAIProvider = (providerId, data) =>
  orderFoodRequest({
    path: `/ai-providers/${encodePath(providerId)}`,
    method: 'put',
    data,
    mutation: true
  })

export const updateAIProviderStatus = (providerId, data) =>
  orderFoodRequest({
    path: `/ai-providers/${encodePath(providerId)}/status`,
    method: 'put',
    data,
    mutation: true
  })

export const deleteAIProvider = (providerId, data) =>
  orderFoodRequest({
    path: `/ai-providers/${encodePath(providerId)}`,
    method: 'delete',
    data,
    mutation: true
  })

export const testAIProviderConnection = (providerId, data) =>
  orderFoodRequest({
    path: `/ai-providers/${encodePath(providerId)}/connection-tests`,
    method: 'post',
    data,
    mutation: true
  })

// AI 模型
export const getAIModelList = (params) =>
  orderFoodRequest({
    path: '/ai-models',
    method: 'get',
    params
  })

export const getAIModelDetail = (modelId) =>
  orderFoodRequest({
    path: `/ai-models/${encodePath(modelId)}`,
    method: 'get'
  })

export const createAIModel = (data) =>
  orderFoodRequest({
    path: '/ai-models',
    method: 'post',
    data,
    mutation: true
  })

export const updateAIModel = (modelId, data) =>
  orderFoodRequest({
    path: `/ai-models/${encodePath(modelId)}`,
    method: 'put',
    data,
    mutation: true
  })

export const updateAIModelStatus = (modelId, data) =>
  orderFoodRequest({
    path: `/ai-models/${encodePath(modelId)}/status`,
    method: 'put',
    data,
    mutation: true
  })

export const deleteAIModel = (modelId, data) =>
  orderFoodRequest({
    path: `/ai-models/${encodePath(modelId)}`,
    method: 'delete',
    data,
    mutation: true
  })

// 六项 AI 执行能力与提示词
export const getAICapabilityList = (params) =>
  orderFoodRequest({
    path: '/ai-capabilities',
    method: 'get',
    params
  })

export const getAICapabilityDetail = (capabilityCode) =>
  orderFoodRequest({
    path: `/ai-capabilities/${encodePath(capabilityCode)}`,
    method: 'get'
  })

export const updateAICapability = (capabilityCode, data) =>
  orderFoodRequest({
    path: `/ai-capabilities/${encodePath(capabilityCode)}`,
    method: 'put',
    data,
    mutation: true
  })

export const getAICapabilityPromptWorkspace = (capabilityCode) =>
  orderFoodRequest({
    path: `/ai-capabilities/${encodePath(capabilityCode)}/prompts`,
    method: 'get'
  })

export const updateAICapabilityPrompt = (capabilityCode, data) =>
  orderFoodRequest({
    path: `/ai-capabilities/${encodePath(capabilityCode)}/prompts`,
    method: 'put',
    data,
    mutation: true
  })

export const validateAICapabilityPrompt = (capabilityCode, data) =>
  orderFoodRequest({
    path: `/ai-capabilities/${encodePath(capabilityCode)}/prompt-validation`,
    method: 'post',
    data,
    mutation: true
  })

export const renderAICapabilityPrompt = (capabilityCode, data) =>
  orderFoodRequest({
    path: `/ai-capabilities/${encodePath(capabilityCode)}/prompt-render-preview`,
    method: 'post',
    data,
    mutation: true
  })

export const testAICapabilityPrompt = (capabilityCode, data) =>
  orderFoodRequest({
    path: `/ai-capabilities/${encodePath(capabilityCode)}/prompt-tests`,
    method: 'post',
    data,
    mutation: true
  })
