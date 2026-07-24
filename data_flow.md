# Data Flow Diagrams

## Text File Parsing

```mermaid
flowchart TD
    A[Start] --> B[Read Text File]
    B --> C[Parse Range]
    C --> D[Validate Range]
    D --> E[Extract Content]
    E --> F[Copy to Clipboard]
    F --> G[End]
```

## Markdown File Parsing

```mermaid
flowchart TD
    A[Start] --> B[Read Markdown File]
    B --> C[Parse Range]
    C --> D[Validate Range]
    D --> E[Extract Content]
    E --> F[Copy to Clipboard]
    F --> G[End]
```

## JSON File Parsing

```mermaid
flowchart TD
    A[Start] --> B[Read JSON File]
    B --> C[Parse Range]
    C --> D[Validate Range]
    D --> E[Extract Content]
    E --> F[Copy to Clipboard]
    F --> G[End]