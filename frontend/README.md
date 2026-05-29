# PentAGI Frontend

A chat application built with React, TypeScript, and GraphQL that enables intelligent conversations with AI agents.

## Features

-   💬 Real-time chat interface with AI agents
-   🤖 Multiple AI agent support and management
-   📊 Real-time terminal output monitoring
-   🎯 Task and subtask tracking system
-   🔍 Integrated search capabilities
-   📚 Vector store for knowledge base management
-   📸 Screenshot capture and management
-   🌓 Dark/Light theme support
-   📱 Responsive design (mobile, tablet, desktop)
-   🔐 Authentication system with multiple providers
-   🔄 Real-time updates via GraphQL subscriptions
-   ⚡ High-performance React components

## Tech Stack

-   **Framework**: React 18 with TypeScript
-   **Build Tool**: Vite
-   **Styling**: Tailwind CSS
-   **UI Components**:
    -   shadcn/ui
    -   Radix UI primitives
    -   Lucide icons
-   **State Management**:
    -   React Context
    -   Custom Hooks
-   **API Integration**:
    -   GraphQL
    -   Apollo Client
    -   WebSocket subscriptions
-   **Type Safety**: TypeScript
-   **Authentication**: Multiple provider support
-   **Code Quality**:
    -   ESLint
    -   Prettier
    -   TypeScript strict mode

## Project Structure

src/
├── components/ # Shared UI components
│ ├── ui/ # Base UI components
│ └── icons/ # SVG icons and logo
├── features/ # Feature-based modules
│ ├── chat/ # Chat related components
│ ├── authentication/ # Auth related components
├── hooks/ # Custom React hooks
├── lib/ # Utilities and configurations
├── graphql/ # GraphQL operations and types
├── models/ # TypeScript interfaces
└── pages/ # Application routes

## Key Components

### Chat Interface

-   Split view with messages and tools panels
-   Resizable panels for desktop
-   Mobile-optimized view with tabs
-   Real-time message updates

### Task System

-   Real-time task tracking
-   Subtask management
-   Progress monitoring
-   Status updates

### Terminal

-   Command output display
-   Real-time updates
-   Scrollable history
-   Syntax highlighting

### Vector Store

-   Knowledge base integration
-   Search capabilities
-   Data management

### Agent System

-   Multi-agent support
-   Agent status monitoring
-   Agent communication logs

## Development

### Prerequisites

-   Node.js 18+
-   pnpm 10+

### Installation

1. Clone the repository
2. Install dependencies:
   pnpm install
3. Start the development server:
   pnpm run dev

### Building for Production

pnpm run build

### Environment Variables

Create a .env file in the root directory:

VITE_API_URL=your_api_url

## Contributing

1. Fork the repository
2. Create your feature branch (git checkout -b feature/amazing-feature)
3. Commit your changes (git commit -m 'Add some amazing feature')
4. Push to the branch (git push origin feature/amazing-feature)
5. Open a Pull Request
