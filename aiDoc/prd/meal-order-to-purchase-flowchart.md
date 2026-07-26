# 饭局点餐到采购清单流程图

本文档描述从创建饭局、参与者点菜、关闭点餐、确认最终菜单、生成采购清单到创建者手动完成饭局的主链路。

适用口径：

- 饭局主状态为 `collecting`、`closed`、`confirmed`、`completed`、`cancelled`。
- 点餐码过期不作为独立主状态，而是 `closed` 的关闭原因。
- 点餐阶段不保存完整菜品快照，确认最终菜单时才生成饭局菜品快照。
- 采购清单不维护采购主状态，清单项只区分未购买和已购买。
- AI 备菜提醒是确认菜单后的可选能力，发起一次按后台配置消耗积分。

```mermaid
%%{init: {"theme":"base","themeVariables":{"fontFamily":"Inter, PingFang SC, Microsoft YaHei, sans-serif","primaryColor":"#F8FAFC","primaryTextColor":"#0F172A","primaryBorderColor":"#CBD5E1","lineColor":"#64748B"},"flowchart":{"curve":"linear","nodeSpacing":44,"rankSpacing":48,"padding":14}}}%%
flowchart LR
    subgraph P1["1. 创建饭局"]
        direction TB
        A([创建饭局]):::start
        B[填写名称和有效期<br/>选择候选菜]:::action
        C[生成点餐码<br/>状态 collecting]:::state
        A --> B --> C
    end

    subgraph P2["2. 参与点菜"]
        direction TB
        D{点餐码可加入?}:::decision
        D1[提示不可加入<br/>无效 / 已关闭 / 已取消]:::error
        E[创建或复用参与记录]:::system
        F[点选想吃菜品<br/>同菜仅保留一条记录]:::action
        G[统计想吃人数]:::system
        D -->|否| D1
        D -->|是| E --> F --> G
    end

    subgraph P3["3. 关闭与确认"]
        direction TB
        H{点餐是否结束?}:::decision
        H1[保持 collecting<br/>继续等待点菜]:::muted
        I[状态 closed<br/>原因：手动关闭]:::state
        J[状态 closed<br/>原因：过期关闭]:::state
        K([状态 cancelled<br/>流程终止]):::stop
        L[创建者查看统计结果]:::action
        M[按基础份量<br/>建议制作份数]:::system
        N[确认最终菜单<br/>和制作份数]:::action
        H -->|否| H1
        H -->|手动关闭| I --> L
        H -->|到期| J --> L
        H -->|取消| K
        L --> M --> N
    end

    subgraph P4["4. 采购与提醒"]
        direction TB
        O[生成饭局菜品快照<br/>固定菜名 / 做法 / 配料 / 最终份数]:::system
        P[生成采购清单<br/>按配料名 + 单位合并数量]:::system
        Q([状态 confirmed<br/>采购清单可用]):::complete
        R[维护采购清单<br/>未购买 / 已购买<br/>编辑 / 新增 / 删除 / 分享]:::action
        S{发起 AI 备菜提醒?}:::decision
        T[扣积分并生成<br/>优先级和颜色提醒]:::ai
        U[创建者手动完成饭局]:::action
        V([状态 completed<br/>可创建下一场]):::complete
        O --> P --> Q --> R --> S
        S -->|否| U --> V
        S -->|是| T --> U
    end

    C --> D
    G --> H
    N --> O

    style P1 fill:#F8FAFC,stroke:#CBD5E1,stroke-width:1px,color:#334155;
    style P2 fill:#F8FAFC,stroke:#CBD5E1,stroke-width:1px,color:#334155;
    style P3 fill:#F8FAFC,stroke:#CBD5E1,stroke-width:1px,color:#334155;
    style P4 fill:#F8FAFC,stroke:#CBD5E1,stroke-width:1px,color:#334155;

    classDef start fill:#DCFCE7,stroke:#16A34A,color:#14532D,stroke-width:1.5px;
    classDef complete fill:#E0F2FE,stroke:#0284C7,color:#0C4A6E,stroke-width:1.5px;
    classDef stop fill:#FEE2E2,stroke:#DC2626,color:#7F1D1D,stroke-width:1.5px;
    classDef action fill:#FFFFFF,stroke:#64748B,color:#0F172A,stroke-width:1.2px;
    classDef system fill:#EFF6FF,stroke:#2563EB,color:#1E3A8A,stroke-width:1.2px;
    classDef decision fill:#FFF7ED,stroke:#F97316,color:#7C2D12,stroke-width:1.4px;
    classDef state fill:#F1F5F9,stroke:#475569,color:#0F172A,stroke-width:1.2px;
    classDef ai fill:#F5F3FF,stroke:#7C3AED,color:#4C1D95,stroke-width:1.2px;
    classDef error fill:#FEF2F2,stroke:#EF4444,color:#7F1D1D,stroke-width:1.2px;
    classDef muted fill:#F8FAFC,stroke:#94A3B8,color:#475569,stroke-width:1px;
```

边界说明：

- 候选源菜品被删除或不可用时，当前饭局保留候选记录并标记“源菜品已失效”，但不可继续点选、已有票数不计入统计且不能进入最终菜单。
- 创建者从候选菜单移除菜品时，参与者端不再展示，相关点菜记录不计入最终菜单。
- 点餐阶段不保存完整菜品快照，只有创建者确认最终菜单和制作份数后才生成饭局菜品快照。
- 采购清单不维护采购主状态，整体进度由清单项“未购买 / 已购买”推导。
- AI 备菜提醒不作为强制计划，用户可手动调整、忽略或重新生成，不影响饭局、最终菜单和采购清单。
- `confirmed` 仍是进行中饭局；只有创建者手动进入 `completed` 后才释放“只能创建一个进行中饭局”的约束。
