---
# 注意不要修改本文头文件，如修改，CodeBuddy（内网版）将按照默认逻辑设置
type: always
---

# GitHub Copilot Code Review Instructions

This document provides comprehensive code review guidelines for the **bkmonitor-ecosystem** project.

The **bkmonitor-ecosystem** project provides "minimalist" out-of-the-box observability data ingestion demos to help users quickly get started with BlueKing Observability Platform functions. It covers multiple languages including Go, Python, Java, JavaScript, and C++.

[TOC]

## Core Review Principles

As a Copilot agent reviewing this project, prioritize the following principles:

1.  **Clarity and Maintainability**: Code must be easy to read and understand.
2.  **Best Practices**: Follow industry-standard coding conventions and observability best practices.
3.  **Correctness and Reliability**: Ensure demos function correctly and reliably in a Docker environment.
4.  **MANDATORY Review Acknowledgment**: This is a **REQUIRED** and **NON-NEGOTIABLE** rule. **All code reviews must be provided in Chinese.** When GitHub Copilot generates any PR review summary, it **MUST** begin with the following exact acknowledgment statement:
    > ✓ 代码评审按照 `.github/copilot-instructions.md` 中定义的指南进行。

    **Important Notes:**
    - This statement confirms that ALL guidelines in this document have been actively applied.
    - Omitting this acknowledgment statement is **NOT ACCEPTABLE**.
    - The statement must appear at the **very beginning** of every PR review summary.

---

## Documentation Review Guidelines

### Chinese Documentation

Chinese documentation must follow the [Chinese Copywriting Guidelines](https://github.com/ruanyf/document-style-guide).

**Key Requirements:**

- **Spacing**:
  - Add space between Chinese and English characters
  - Add space between Chinese and numbers
  - Add space between numbers and units (except for degrees, percentages)
- **Punctuation**:
  - Use full-width punctuation for Chinese text
  - Use half-width punctuation for English text and code
  - Do not add space before or after full-width punctuation
- **Nouns**:
  - Use correct capitalization for proper nouns (e.g., GitHub, Python, macOS)
  - Maintain brand name capitalization
- **Numbers**:
  - Use Arabic numerals for statistics and measurements
  - Numbers with more than 4 digits should use comma separators (e.g., 1,000)
- **Links**: Ensure link text is meaningful and descriptive

### Code Examples in Documentation

Code examples in documentation **MUST** strictly follow this structure:
1.  **Explanation**: A clear description of what the code does.
2.  **Minimal Code Snippet**: A concise, runnable code block focused on the specific concept.
3.  **Reference Link**: A link to the official documentation (e.g., OpenTelemetry docs).

**Example (Good):**
> Attributes（属性）是 Span 元数据，以 Key-Value 形式存在。
>
> 在 Span 设置属性，对问题定位、过滤、聚合非常有帮助。
>
> ```go
> // 增加 Span 自定义属性
> span.SetAttributes(
>     attribute.Int("helloworld.kind", 1),
>     attribute.String("helloworld.step", "tracesCustomSpanDemo"),
> )
> ```
>
> * <a href="https://opentelemetry.io/docs/languages/go/instrumentation/#span-attributes" target="_blank">Span Attributes</a>

---

## Contribution Guidelines (Project Specific)

**CRITICAL**: Ensure adherence to the project's specific contribution workflow defined in `README.md`.

1.  **Demo Structure**:
    - Demos that cannot be open-sourced must be placed under `inner` directories (e.g., `examples/go-examples/inner`).
    - **Docker Support**: All demos **MUST** provide a `Docker` startup method.

2.  **Documentation Rendering**:
    - **Template-Based**: Documents intended for external open-source must be written as templates under the `templates/` directory.
    - **Avoid Direct Edits**: Do not directly edit files in `docs/` if they are generated from templates.
    - **Rendering Check**: If documentation is added or modified in `templates/`, verify that `make render` has been executed to generate the corresponding files in `docs/`.
    - **Reminder**: If a PR adds documentation but fails to follow the `templates/` -> `docs/` rendering workflow, **issue a mandatory reminder** to the author.

3.  **Domain Desensitization**:
    - **No Internal Domains**: Files in `templates/` and `examples/` (excluding `inner` subdirectories) **MUST NOT** contain internal domains (e.g., `woa.com`).
    - **Use Variables**: Use variables defined in `tools/formatter/context/inner.py` or define new ones there to handle internal URLs.
    - **Example**: Use `{{ECOSYSTEM_REPOSITORY_URL}}` instead of the hardcoded internal Git URL.

---

## Multi-Language Code Review Guidelines

Code in this project spans multiple languages. Adhere to **Google Style Guides** for all languages.

### 1. General Coding Standards

- **Go**: Follow the [Google Go Style Guide](https://google.github.io/styleguide/go/).
- **Python**: Follow the [Google Python Style Guide](https://google.github.io/styleguide/pyguide.html) and PEP 8.
- **Java**: Follow the [Google Java Style Guide](https://google.github.io/styleguide/javaguide.html).
- **JavaScript**: Follow the [Google JavaScript Style Guide](https://google.github.io/styleguide/jsguide.html).
- **C++**: Follow the [Google C++ Style Guide](https://google.github.io/styleguide/cppguide.html).

### 2. Critical Configuration Comments

For critical reporting configuration parameters (e.g., Endpoint, API_URL, TOKEN), you **MUST** add conspicuous comments ❗❗【非常重要】 to alert developers.

**Example (Good):**
```go
return newHttpTracerExporter(
    ctx,
    // ❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
    // 格式为 ip:port 或 domain:port，不要带 schema
    s.config.Endpoint,
    // ❗❗【非常重要】请传入应用 Token
    map[string]string{"x-bk-token": s.config.Token},
)
```

### 3. Observability Standards

This project focuses on observability. Code and configuration must adhere to industry standards and best practices for metrics, traces, and logs.

- **OpenTelemetry**:
  - Follow [OpenTelemetry Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/) for resource, trace, and metric naming.
  - Ensure correct propagation of context (Trace Context) across service boundaries.
- **Prometheus**:
  - Follow [Prometheus Metric and Label Naming Best Practices](https://prometheus.io/docs/practices/naming/).
  - Use appropriate metric types (Counter, Gauge, Histogram, Summary).
  - Avoid high-cardinality labels unless necessary.

---

## Review Tone & Communication

- **Be Constructive**: Focus on improving the code, not criticizing the author.
- **Be Specific**: Provide concrete examples and suggestions.
- **Be Educational**: Explain the reasoning behind suggestions, especially regarding observability best practices.
- **Language**: Provide all feedback in **Chinese**.

---

## References

- [Chinese Copywriting Guidelines](https://github.com/ruanyf/document-style-guide)
- [Google Style Guides](https://google.github.io/styleguide/)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)

---

**Last Updated**: 2025-12-14
