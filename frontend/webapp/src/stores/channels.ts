import { atom } from "nanostores"
import { UnaryInterceptor } from "grpc-web";
import { $authToken } from "../auth/store/auth";
import { $currentTenant } from "./tenants";
import { 
  ChannelServiceClient, 
  Channel, 
  CreateChannelRequest, 
  UpdateChannelRequest, 
  GetChannelsRequest,
  GetChannelMembersRequest,
  AddChannelMemberRequest,
  RemoveChannelMemberRequest,
  ChannelMember
} from "../proto/videoservice"

export const $channels = atom<Channel[]>([])
export const $channelMembers = atom<ChannelMember[]>([])
export const $currentChannel = atom<Channel | null>(null)
export const $isLoadingChannels = atom(false)
export const $channelError = atom<string | null>(null)

const unaryInterceptor: UnaryInterceptor<any, any> = {
  intercept: (request, invoker) => {
    const m = request.getMetadata();
    const token = $authToken.get();
    const currentTenant = $currentTenant.get();
    
    m["authorization"] = token;
    
    // Add tenant ID header if available
    if (currentTenant?.tenant?.id) {
      m["x-tenant-id"] = currentTenant.tenant.id;
    }
    
    return invoker(request);
  },
};

export const channelService = new ChannelServiceClient(
  import.meta.env.VITE_PUBLIC_API_URL.replace(/\/$/, ""),
  {},
  {
    unaryInterceptors: [unaryInterceptor],
  }
);

// Fetch all channels for current tenant
export const fetchChannels = async () => {
  try {
    $isLoadingChannels.set(true)
    $channelError.set(null)
    
    const request = new GetChannelsRequest()
    const response = await channelService.GetChannels(request, {})
    
    if (response.channels) {
      $channels.set(response.channels)
    } else {
      $channelError.set(response.message || 'Failed to fetch channels')
      $channels.set([])
    }
  } catch (error) {
    console.error("Error fetching channels:", error)
    $channelError.set('Failed to fetch channels')
    $channels.set([])
    throw error
  } finally {
    $isLoadingChannels.set(false)
  }
}

// Get channel from existing list by ID
export const getChannelById = (channelId: string): Channel | null => {
  const channels = $channels.get()
  const channel = channels.find(ch => ch.id === channelId) || null
  $currentChannel.set(channel)
  return channel
}

// Create a new channel
export const createChannel = async (name: string, description?: string) => {
  try {
    $channelError.set(null)
    
    const request = new CreateChannelRequest({
      name,
      description: description || ''
    })
    
    const response = await channelService.CreateChannel(request, {})
    
    if (response.channel) {
      // Add the new channel to the existing list
      const currentChannels = $channels.get()
      $channels.set([...currentChannels, response.channel])
      return response.channel
    } else {
      $channelError.set(response.message || 'Failed to create channel')
      return null
    }
  } catch (error) {
    console.error("Error creating channel:", error)
    $channelError.set('Failed to create channel')
    return null
  }
}

// Update an existing channel
export const updateChannel = async (channelId: string, name: string, description?: string) => {
  try {
    $channelError.set(null)
    
    const request = new UpdateChannelRequest({
      channel_id: channelId,
      name,
      description: description || ''
    })
    
    const response = await channelService.UpdateChannel(request, {})
    
    if (response.channel) {
      // Update the channel in the list
      const currentChannels = $channels.get()
      const updatedChannels = currentChannels.map(ch => 
        ch.id === channelId ? response.channel! : ch
      )
      $channels.set(updatedChannels)
      $currentChannel.set(response.channel)
      return response.channel
    } else {
      $channelError.set(response.message || 'Failed to update channel')
      return null
    }
  } catch (error) {
    console.error("Error updating channel:", error)
    $channelError.set('Failed to update channel')
    return null
  }
}

// Fetch channel members
export const fetchChannelMembers = async (channelId: string): Promise<ChannelMember[]> => {
  try {
    const request = new GetChannelMembersRequest()
    request.channel_id = channelId

    const response = await channelService.GetMembers(request, {})
    
    if (response.channel_members) {
      $channelMembers.set(response.channel_members)
      return response.channel_members
    } else {
      $channelMembers.set([])
      return []
    }
  } catch (error: any) {
    console.error('Failed to fetch channel members:', error)
    $channelMembers.set([])
    throw new Error(error.message || 'Failed to fetch channel members')
  }
}

// Add channel member
export const addChannelMember = async (channelId: string, userId: string, role: string): Promise<void> => {
  try {
    const request = new AddChannelMemberRequest({
      channel_id: channelId,
      user_id: userId,
      role: role
    })

    await channelService.AddMember(request, {})
  } catch (error: any) {
    console.error('Failed to add channel member:', error)
    throw new Error(error.message || 'Failed to add channel member')
  }
}

// Remove channel member
export const removeChannelMember = async (channelId: string, userId: string): Promise<void> => {
  try {
    const request = new RemoveChannelMemberRequest({
      channel_id: channelId,
      user_id: userId
    })

    await channelService.RemoveMember(request, {})
  } catch (error: any) {
    console.error('Failed to remove channel member:', error)
    throw new Error(error.message || 'Failed to remove channel member')
  }
}

// Auto-fetch channels when tenant changes
$currentTenant.subscribe((tenant) => {
  if (tenant?.tenant?.id) {
    fetchChannels()
  } else {
    $channels.set([])
    $channelMembers.set([])
    $currentChannel.set(null)
  }
}) 