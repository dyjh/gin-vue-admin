# 饭局点餐到采购清单流程图

本文档描述从创建饭局、参与者点菜、关闭点餐、确认最终菜单到生成采购清单的主链路。

适用口径：

- 饭局主状态为 `collecting`、`closed`、`confirmed`、`cancelled`。
- 点餐码过期不作为独立主状态，而是 `closed` 的关闭原因。
- 点餐阶段不保存完整菜品快照，确认最终菜单时才生成饭局菜品快照。
- 采购清单不维护采购主状态，清单项只区分未购买和已购买。
- AI 做菜规划是确认菜单后的可选能力，发起一次按后台配置消耗积分。

```mermaid
%%{init: {"theme":"base","themeVariables":{"fontFamily":"Inter, PingFang SC, Microsoft YaHei, sans-serif","primaryColor":"#F8FAFC","primaryTextColor":"#0F172A","primaryBorderColor":"#CBD5E1","lineColor":"#64748B","secondaryColor":"#ECFDF5","tertiaryColor":"#FFF7ED","noteBkgColor":"#FFFBEB","noteTextColor":"#713F12"},"flowchart":{"curve":"basis","nodeSpacing":48,"rankSpacing":58,"padding":16}}}%%
flowchart TD
    start([创建者发起饭局]):::start

    subgraph creator["创建者操作"]
        c1[填写饭局名称<br/>设置点餐码有效期]:::actor
        c2[从可用菜品或菜谱<br/>选择候选菜]:::actor
        c3[提前关闭点餐]:::actor
        c4[查看点菜统计]:::actor
        c5[确认最终菜单<br/>和制作份数]:::actor
        c6[维护采购清单<br/>勾选 / 编辑 / 新增 / 删除 / 分享]:::actor
    end

    subgraph participant["参与者操作"]
        p1[微信登录<br/>输入点餐码]:::actor
        p2[进入饭局点菜页]:::actor
        p3[点选或取消想吃菜品]:::actor
    end

    subgraph system["系统处理"]
        s1[创建饭局<br/>状态 collecting<br/>生成点餐码]:::system
        s2{点餐码是否可加入}:::decision
        s3{是否已加入该饭局}:::decision
        s4[创建或复用参与记录]:::system
        s5[更新点菜记录<br/>同一参与者同一候选菜只保留一条]:::system
        s6{关闭点餐触发}:::decision
        s7[状态 closed<br/>关闭原因 手动关闭]:::state
        s8[状态 closed<br/>关闭原因 过期关闭]:::state
        s9[统计每道菜想吃人数]:::system
        s10[按基础份量<br/>建议制作份数]:::system
        s11[生成饭局菜品快照<br/>菜名 / 做法 / 配料 / 最终份数]:::system
        s12[生成采购清单<br/>按配料名 + 单位合并数量]:::system
        s13[状态 confirmed<br/>采购清单可用]:::complete
    end

    subgraph optional["可选能力"]
        a1{是否发起 AI 做菜规划}:::decision
        a2[校验能力开关 / 积分 / 限额<br/>生成做菜顺序和备菜建议]:::optional
    end

    invalid([提示不可加入<br/>无效 / 已取消 / 非收集中]):::error
    cancelled([状态 cancelled<br/>不再点餐或确认菜单]):::stop
    done([流程结束]):::complete

    note1[[候选菜边界<br/>菜品被下架：当前饭局候选项保留并标记已下架<br/>创建者移除候选菜：参与者端不展示，统计不计入最终菜单]]:::note
    note2[[采购清单边界<br/>不维护采购主状态<br/>每个清单项只在未购买 / 已购买之间切换]]:::note

    start --> c1 --> c2 --> s1
    c2 -.-> note1

    s1 --> p1 --> s2
    s2 -->|否| invalid
    s2 -->|是| s3
    s3 -->|已加入| s4
    s3 -->|未加入| s4
    s4 --> p2 --> p3 --> s5
    s5 --> s6

    s1 --> s6
    s6 -->|继续收集| p1
    s6 -->|创建者手动关闭| c3 --> s7
    s6 -->|点餐码到期| s8
    s6 -->|创建者取消| cancelled

    s7 --> c4
    s8 --> c4
    s7 -.->|创建者取消| cancelled
    s8 -.->|创建者取消| cancelled
    c4 --> s9 --> s10 --> c5 --> s11 --> s12 --> s13
    s12 -.-> note2

    s13 --> c6
    s13 --> a1
    a1 -->|否| done
    a1 -->|是| a2 --> done

    classDef start fill:#DCFCE7,stroke:#16A34A,color:#14532D,stroke-width:1.5px;
    classDef complete fill:#E0F2FE,stroke:#0284C7,color:#0C4A6E,stroke-width:1.5px;
    classDef stop fill:#FEE2E2,stroke:#DC2626,color:#7F1D1D,stroke-width:1.5px;
    classDef actor fill:#F8FAFC,stroke:#64748B,color:#0F172A,stroke-width:1.2px;
    classDef system fill:#EFF6FF,stroke:#2563EB,color:#1E3A8A,stroke-width:1.2px;
    classDef decision fill:#FFF7ED,stroke:#F97316,color:#7C2D12,stroke-width:1.4px;
    classDef state fill:#F1F5F9,stroke:#475569,color:#0F172A,stroke-width:1.2px;
    classDef optional fill:#F5F3FF,stroke:#7C3AED,color:#4C1D95,stroke-width:1.2px;
    classDef error fill:#FEF2F2,stroke:#EF4444,color:#7F1D1D,stroke-width:1.2px;
    classDef note fill:#FFFBEB,stroke:#D97706,color:#713F12,stroke-dasharray:5 5;
```
