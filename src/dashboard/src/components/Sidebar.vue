<script setup lang="ts">
import { ref, inject } from 'vue'

const isConnected = inject<any>('isConnected')
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar-logo">
      <img src="/logo.svg" alt="OFA Logo" style="width: 40px; height: 40px; margin-bottom: 0.5rem;">
      <h1>OFA</h1>
      <span>Omni Federated Agents</span>
    </div>

    <nav>
      <ul class="sidebar-nav">
        <li>
          <router-link to="/">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="7" height="7"></rect>
              <rect x="14" y="3" width="7" height="7"></rect>
              <rect x="14" y="14" width="7" height="7"></rect>
              <rect x="3" y="14" width="7" height="7"></rect>
            </svg>
            Dashboard
          </router-link>
        </li>
        <li>
          <router-link to="/agents">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
              <circle cx="9" cy="7" r="4"></circle>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
            </svg>
            Agents
          </router-link>
        </li>
        <li>
          <router-link to="/tasks">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
              <polyline points="14 2 14 8 20 8"></polyline>
              <line x1="16" y1="13" x2="8" y2="13"></line>
              <line x1="16" y1="17" x2="8" y2="17"></line>
              <polyline points="10 9 9 9 8 9"></polyline>
            </svg>
            Tasks
          </router-link>
        </li>
        <li>
          <router-link to="/monitor">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="18" y1="20" x2="18" y2="10"></line>
              <line x1="12" y1="20" x2="12" y2="4"></line>
              <line x1="6" y1="20" x2="6" y2="14"></line>
            </svg>
            Monitor
          </router-link>
        </li>
        <li>
          <router-link to="/messages">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
            </svg>
            Messages
          </router-link>
        </li>
      </ul>
    </nav>

    <div style="position: absolute; bottom: 1rem; left: 1.5rem; right: 1.5rem;">
      <div style="display: flex; align-items: center; gap: 0.5rem; font-size: 0.75rem; color: var(--gray-400);">
        <span :class="isConnected ? 'badge-success' : 'badge-danger'" style="width: 8px; height: 8px; border-radius: 50%;"></span>
        {{ isConnected ? 'Connected' : 'Disconnected' }}
      </div>
    </div>
  </aside>
</template>