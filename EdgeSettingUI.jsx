import { useState } from "react";

/* ━━ 设计 Token ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
const T = {
  bg:           '#f0f5ff',
  bgCard:       '#ffffff',
  primary:      '#42aaf5',
  primaryDark:  '#1a85d0',
  primarySoft:  'rgba(66,170,245,0.10)',
  primaryBorder:'rgba(66,170,245,0.30)',
  border:       'rgba(66,170,245,0.20)',
  borderSubtle: 'rgba(66,170,245,0.12)',
  shadow:       '0 2px 12px rgba(66,170,245,0.13)',
  shadowSm:     '0 1px 6px rgba(66,170,245,0.10)',
  text:         '#1a1c3a',
  subText:      '#5b6080',
  muted:        '#9ca3af',
  success:      '#10b981',
  successSoft:  'rgba(16,185,129,0.10)',
  successBorder:'rgba(16,185,129,0.28)',
  warning:      '#f59e0b',
  warningSoft:  'rgba(245,158,11,0.10)',
  danger:       '#ef4444',
  dangerSoft:   'rgba(239,68,68,0.08)',
  dangerBorder: 'rgba(239,68,68,0.24)',
  system:       '#8b5cf6',
  // 圆角
  r8:  '8px',
  r12: '12px',
  r14: '14px',
  r20: '20px',
};

const CSS = `
@import url('https://fonts.googleapis.com/css2?family=DM+Sans:ital,opsz,wght@0,9..40,400;0,9..40,600;0,9..40,700&family=Noto+Sans+SC:wght@400;500;700&display=swap');
*{box-sizing:border-box;margin:0;padding:0;}
@keyframes blink{0%,100%{opacity:1}50%{opacity:.2}}
`;

/* ━━ 原子组件 ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
function Ring({ pct, color, size = 44 }) {
  const r = 14, c = 2 * Math.PI * r;
  return (
    <svg width={size} height={size} viewBox="0 0 36 36" style={{ flexShrink: 0 }}>
      <circle cx="18" cy="18" r={r} fill="none" stroke={`${color}20`} strokeWidth="4.5"/>
      <circle cx="18" cy="18" r={r} fill="none" stroke={color} strokeWidth="4.5"
        strokeDasharray={`${(pct/100)*c} ${c}`} strokeLinecap="round"
        transform="rotate(-90 18 18)"/>
    </svg>
  );
}

function Dot({ status }) {
  const c = { running: T.success, stopped: T.muted }[status] ?? T.muted;
  return <div style={{ width:8, height:8, borderRadius:'50%', background:c, flexShrink:0,
    animation: status==='running'?'blink 2.4s infinite':'none' }}/>;
}

function LvBadge({ lv }) {
  const map = { INFO:{t:'信息',c:T.success,bg:T.successSoft}, WARN:{t:'警告',c:T.warning,bg:T.warningSoft}, ERROR:{t:'错误',c:T.danger,bg:T.dangerSoft} };
  const s = map[lv] ?? map.INFO;
  return <span style={{ fontSize:11, fontWeight:700, padding:'2px 7px', borderRadius:6, background:s.bg, color:s.c }}>{s.t}</span>;
}

/* ━━ 布局外壳 ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
// 窄屏（手机）模拟器
function NarrowFrame({ children, label }) {
  return (
    <div style={{ display:'flex', flexDirection:'column', alignItems:'center', gap:10 }}>
      <div style={{ width:320, height:650, background:T.bg, borderRadius:44,
        border:`1.5px solid ${T.borderSubtle}`,
        boxShadow:`0 0 0 1.5px #fff, 0 20px 56px rgba(66,170,245,0.14)`,
        position:'relative', overflow:'hidden', fontFamily:"'DM Sans','Noto Sans SC',sans-serif" }}>
        {/* 刘海 */}
        <div style={{ position:'absolute', top:10, left:'50%', transform:'translateX(-50%)',
          width:78, height:20, background:'rgba(240,245,255,0.95)', borderRadius:10, zIndex:20,
          border:`1px solid ${T.borderSubtle}` }}/>
        {/* 状态栏 */}
        <div style={{ position:'absolute', top:0, left:0, right:0, height:46, zIndex:15,
          display:'flex', alignItems:'flex-end', justifyContent:'space-between',
          padding:'0 20px 7px', fontSize:11, color:T.subText, fontWeight:600 }}>
          <span>09:41</span>
          <div style={{ display:'flex', gap:4, alignItems:'center' }}>
            <svg width="14" height="11" viewBox="0 0 13 10">
              {[0,3,6,9].map((x,i)=><rect key={x} x={x} y={9-3-i*1.8} width="2.5" height={3+i*1.8} rx=".8"
                fill={i<3?T.primary:'rgba(66,170,245,0.22)'}/>)}
            </svg>
            <svg width="16" height="12" viewBox="0 0 15 11">
              <rect x="1" y="2" width="11" height="8" rx="2" fill="none" stroke={T.primaryDark} strokeWidth="1.2" opacity=".5"/>
              <rect x="2.5" y="3.5" width="7.5" height="5" rx="1" fill={T.primaryDark} opacity=".5"/>
            </svg>
          </div>
        </div>
        <div style={{ position:'absolute', inset:0, paddingTop:46 }}>{children}</div>
      </div>
      <div style={{ fontSize:12, fontWeight:700, color:T.text, letterSpacing:.2 }}>{label}</div>
    </div>
  );
}

// 宽屏（Pad/桌面）模拟器
function WideFrame({ children, label }) {
  return (
    <div style={{ display:'flex', flexDirection:'column', alignItems:'center', gap:10 }}>
      <div style={{ width:540, height:650, background:T.bg, borderRadius:28,
        border:`1.5px solid ${T.borderSubtle}`,
        boxShadow:`0 0 0 1.5px #fff, 0 16px 48px rgba(66,170,245,0.12)`,
        position:'relative', overflow:'hidden', fontFamily:"'DM Sans','Noto Sans SC',sans-serif" }}>
        {/* 底部 Home Bar */}
        <div style={{ position:'absolute', bottom:8, left:'50%', transform:'translateX(-50%)',
          width:120, height:4, background:'rgba(66,170,245,0.14)', borderRadius:2, zIndex:20 }}/>
        {/* 状态栏 */}
        <div style={{ position:'absolute', top:0, left:0, right:0, height:38, zIndex:15,
          display:'flex', alignItems:'flex-end', justifyContent:'space-between',
          padding:'0 24px 6px', fontSize:12, color:T.subText, fontWeight:600 }}>
          <span>09:41</span>
          <svg width="14" height="11" viewBox="0 0 13 10">
            {[0,3,6,9].map((x,i)=><rect key={x} x={x} y={9-3-i*1.8} width="2.5" height={3+i*1.8} rx=".8"
              fill={i<3?T.primary:'rgba(66,170,245,0.22)'}/>)}
          </svg>
        </div>
        <div style={{ position:'absolute', inset:0, paddingTop:38 }}>{children}</div>
      </div>
      <div style={{ fontSize:12, fontWeight:700, color:T.text, letterSpacing:.2 }}>{label} · 宽屏</div>
    </div>
  );
}

/* ━━ 导航组件 ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
const NAV = [
  { id:'home',    label:'首页',   icon:'◈' },
  { id:'monitor', label:'监控',   icon:'◉' },
  { id:'apps',    label:'应用',   icon:'⊞' },
  { id:'about',   label:'系统',   icon:'◎' },
];

function TabBar({ active }) {
  return (
    <div style={{ display:'flex', borderTop:`1px solid ${T.borderSubtle}`,
      background:'rgba(255,255,255,0.96)', padding:'8px 0 12px', flexShrink:0 }}>
      {NAV.map(t => (
        <div key={t.id} style={{ flex:1, display:'flex', flexDirection:'column', alignItems:'center', gap:3 }}>
          <span style={{ fontSize:22, color:active===t.id?T.primary:T.muted, lineHeight:1.2 }}>{t.icon}</span>
          <span style={{ fontSize:11, color:active===t.id?T.primary:T.muted, fontWeight:active===t.id?700:400 }}>{t.label}</span>
        </div>
      ))}
    </div>
  );
}

function Sidebar({ active }) {
  return (
    <div style={{ width:68, background:'rgba(255,255,255,0.97)',
      borderRight:`1px solid ${T.borderSubtle}`, display:'flex',
      flexDirection:'column', padding:'14px 0', flexShrink:0 }}>
      {/* Logo icon only */}
      <div style={{ display:'flex', justifyContent:'center', paddingBottom:18 }}>
        <div style={{ width:36, height:36, borderRadius:10, background:T.primary,
          display:'flex', alignItems:'center', justifyContent:'center',
          boxShadow:`0 3px 10px rgba(66,170,245,0.4)` }}>
          <span style={{ fontSize:18, color:'#fff' }}>⚙</span>
        </div>
      </div>
      {/* 菜单项：图标 + 小标签竖排 */}
      {NAV.map(it => (
        <div key={it.id} style={{
          display:'flex', flexDirection:'column', alignItems:'center', gap:3,
          padding:'10px 0', cursor:'pointer',
          background:active===it.id?T.primarySoft:'transparent',
          borderRight:active===it.id?`3px solid ${T.primary}`:'3px solid transparent',
        }}>
          <span style={{ fontSize:20, color:active===it.id?T.primary:T.muted, lineHeight:1.2 }}>{it.icon}</span>
          <span style={{ fontSize:10, color:active===it.id?T.primary:T.muted,
            fontWeight:active===it.id?700:400 }}>{it.label}</span>
        </div>
      ))}
    </div>
  );
}

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   页面：① 首页
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
const METRICS = [
  {label:'CPU 占用',  pct:62, color:T.primary},
  {label:'内存使用',  pct:74, color:T.system},
  {label:'磁盘使用',  pct:43, color:T.success},
  {label:'网络流入',  pct:28, color:T.warning},
];
const APPS = [
  {id:'edge-setting',   s:'running', cpu:'0.3%', mem:'28MB'},
  {id:'data-collector', s:'running', cpu:'2.1%', mem:'64MB'},
  {id:'modbus-gateway', s:'running', cpu:'1.2%', mem:'45MB'},
  {id:'rtsp-proxy',     s:'stopped', cpu:'—',    mem:'—'},
  {id:'ota-agent',      s:'running', cpu:'0.1%', mem:'15MB'},
];

function HomeContent() {
  return (
    <div style={{ flex:1, overflowY:'auto', padding:'0 14px' }}>
      {/* 下拉提示 */}
      <div style={{ textAlign:'center', padding:'10px 0 6px', fontSize:12, color:T.muted }}>↓ 下拉刷新</div>

      {/* 2×2 指标卡 */}
      <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:10, marginBottom:12 }}>
        {METRICS.map(m => (
          <div key={m.label} style={{ background:T.bgCard, borderRadius:T.r14,
            padding:'14px 14px 12px', border:`1px solid ${T.borderSubtle}`, boxShadow:T.shadow }}>
            {/* 标签 + 环形 */}
            <div style={{ display:'flex', justifyContent:'space-between', alignItems:'flex-start', marginBottom:6 }}>
              <span style={{ fontSize:12, fontWeight:600, color:T.subText, lineHeight:1.3 }}>{m.label}</span>
              <Ring pct={m.pct} color={m.color} size={38}/>
            </div>
            {/* 数值 */}
            <div style={{ fontSize:28, fontWeight:700, color:T.text, lineHeight:1, marginBottom:10 }}>
              {m.pct}<span style={{ fontSize:13, fontWeight:400, color:T.muted }}>%</span>
            </div>
            {/* 进度条 */}
            <div style={{ height:4, borderRadius:3, background:`${m.color}20` }}>
              <div style={{ height:'100%', width:`${m.pct}%`, borderRadius:3, background:m.color }}/>
            </div>
          </div>
        ))}
      </div>

      {/* 微应用状态卡 */}
      <div style={{ background:T.bgCard, borderRadius:T.r14, padding:'14px 16px',
        border:`1px solid ${T.borderSubtle}`, boxShadow:T.shadow, marginBottom:16 }}>
        {/* 卡头 */}
        <div style={{ display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:12 }}>
          <span style={{ fontSize:16, fontWeight:700, color:T.text }}>微应用状态</span>
          <span style={{ fontSize:13, color:T.primary }}>查看全部 →</span>
        </div>
        {/* 统计行：只有运行中 + 已停止 */}
        <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:10, marginBottom:14 }}>
          {[{label:'运行中',count:4,c:T.success,bg:T.successSoft},{label:'已停止',count:1,c:T.danger,bg:T.dangerSoft}].map(s=>(
            <div key={s.label} style={{ background:s.bg, borderRadius:T.r12, padding:'12px 0', textAlign:'center' }}>
              <div style={{ fontSize:28, fontWeight:700, color:s.c }}>{s.count}</div>
              <div style={{ fontSize:12, color:s.c, opacity:.8, marginTop:2 }}>{s.label}</div>
            </div>
          ))}
        </div>
        {/* 应用列表 */}
        {APPS.map((app,i) => (
          <div key={app.id} style={{ display:'flex', alignItems:'center', gap:10,
            padding:'10px 0', borderBottom: i<APPS.length-1?`1px solid ${T.borderSubtle}`:'none' }}>
            <Dot status={app.s}/>
            <span style={{ fontSize:13, color:T.text, fontFamily:'monospace', flex:1,
              overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap' }}>{app.id}</span>
            <span style={{ fontSize:12, color:T.muted }}>{app.cpu}</span>
            <span style={{ fontSize:12, color:T.muted, width:44, textAlign:'right' }}>{app.mem}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function HomePage({ wide=false }) {
  const topBar = (
    <div style={{ padding:'14px 18px 10px', background:'rgba(255,255,255,0.92)',
      borderBottom:`1px solid ${T.borderSubtle}`,
      display:'flex', alignItems:'center', justifyContent:'space-between', flexShrink:0 }}>
      <div>
        <div style={{ fontSize:18, fontWeight:700, color:T.text }}>设置</div>
      </div>
      <div style={{ fontSize:13, color:T.primary, fontWeight:600 }}>运行 14天 7时</div>
    </div>
  );
  if (wide) return (
    <div style={{ height:'100%', display:'flex', overflow:'hidden' }}>
      <Sidebar active="home"/>
      <div style={{ flex:1, display:'flex', flexDirection:'column', overflow:'hidden' }}>
        {topBar}
        <HomeContent/>
      </div>
    </div>
  );
  return (
    <div style={{ height:'100%', display:'flex', flexDirection:'column', overflow:'hidden' }}>
      {topBar}<HomeContent/><TabBar active="home"/>
    </div>
  );
}

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   页面：② 监控
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
const BARS = [22,35,42,38,55,60,58,62,70,65,68,62,74,66,62];

function MonitorContent() {
  return (
    <div style={{ flex:1, overflowY:'auto', padding:'0 14px' }}>
      <div style={{ textAlign:'center', padding:'10px 0 6px', fontSize:12, color:T.muted }}>↓ 下拉刷新</div>
      {/* 大仪表 */}
      <div style={{ background:T.bgCard, borderRadius:T.r14, padding:'18px 16px',
        border:`1px solid ${T.borderSubtle}`, boxShadow:T.shadow,
        display:'flex', alignItems:'center', gap:20, marginBottom:12 }}>
        <div style={{ position:'relative', display:'flex', alignItems:'center', justifyContent:'center' }}>
          <Ring pct={62} color={T.primary} size={84}/>
          <div style={{ position:'absolute', fontSize:18, fontWeight:700, color:T.text }}>62%</div>
        </div>
        <div>
          <div style={{ fontSize:13, color:T.muted, marginBottom:4 }}>CPU 总体占用率</div>
          <div style={{ fontSize:24, fontWeight:700, color:T.text, lineHeight:1, marginBottom:6 }}>62%</div>
          <div style={{ display:'inline-flex', alignItems:'center', gap:5, background:T.successSoft,
            borderRadius:T.r8, padding:'4px 10px' }}>
            <div style={{ width:6, height:6, borderRadius:'50%', background:T.success }}/>
            <span style={{ fontSize:12, color:T.success, fontWeight:600 }}>正常 · 阈值 80%</span>
          </div>
        </div>
      </div>
      {/* 趋势图 */}
      <div style={{ background:T.bgCard, borderRadius:T.r14, padding:'14px 16px',
        border:`1px solid ${T.borderSubtle}`, boxShadow:T.shadow, marginBottom:12 }}>
        <div style={{ display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:12 }}>
          <span style={{ fontSize:14, fontWeight:700, color:T.text }}>历史趋势</span>
          <div style={{ display:'flex', gap:5 }}>
            {['1时','6时','24时'].map((r,i)=>(
              <span key={r} style={{ fontSize:12, padding:'4px 9px', borderRadius:T.r8,
                background:i===0?T.primary:T.primarySoft, color:i===0?'#fff':T.subText, cursor:'pointer' }}>{r}</span>
            ))}
          </div>
        </div>
        <div style={{ display:'flex', alignItems:'flex-end', gap:3, height:68, marginBottom:5 }}>
          {BARS.map((v,i)=>(
            <div key={i} style={{ flex:1, borderRadius:'3px 3px 0 0', height:`${v}%`,
              background:v>65?`${T.warning}bb`:`${T.primary}77` }}/>
          ))}
        </div>
        <div style={{ display:'flex', justifyContent:'space-between' }}>
          <span style={{ fontSize:11, color:T.muted }}>-60分钟</span>
          <span style={{ fontSize:11, color:T.muted }}>现在</span>
        </div>
      </div>
      {/* 核心占用 */}
      <div style={{ background:T.bgCard, borderRadius:T.r14, padding:'14px 16px',
        border:`1px solid ${T.borderSubtle}`, boxShadow:T.shadow, marginBottom:16 }}>
        <div style={{ fontSize:14, fontWeight:700, color:T.text, marginBottom:12 }}>核心占用</div>
        {[['核心 0',72],['核心 1',54],['核心 2',48],['核心 3',38]].map(([lbl,pct])=>(
          <div key={lbl} style={{ marginBottom:10 }}>
            <div style={{ display:'flex', justifyContent:'space-between', marginBottom:4 }}>
              <span style={{ fontSize:13, color:T.subText }}>{lbl}</span>
              <span style={{ fontSize:13, fontWeight:600, color:T.text }}>{pct}%</span>
            </div>
            <div style={{ height:5, borderRadius:3, background:T.primarySoft }}>
              <div style={{ height:'100%', width:`${pct}%`, borderRadius:3,
                background: pct>70?T.warning:T.primary }}/>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function MonitorPage({ wide=false }) {
  const topBar = (
    <div style={{ padding:'14px 18px 10px', background:'rgba(255,255,255,0.92)',
      borderBottom:`1px solid ${T.borderSubtle}`, flexShrink:0 }}>
      <div style={{ display:'flex', alignItems:'center', justifyContent:'space-between', marginBottom:10 }}>
        <span style={{ fontSize:18, fontWeight:700, color:T.text }}>系统监控</span>
        <div style={{ width:9, height:9, borderRadius:'50%', background:T.success, animation:'blink 2s infinite' }}/>
      </div>
      <div style={{ display:'flex', gap:8 }}>
        {['CPU','内存','磁盘','网络'].map((t,i)=>(
          <div key={t} style={{ padding:'6px 13px', borderRadius:T.r20, fontSize:13, fontWeight:600, cursor:'pointer',
            background:i===0?T.primary:T.primarySoft, color:i===0?'#fff':T.subText,
            border:`1px solid ${i===0?T.primary:T.border}` }}>{t}</div>
        ))}
      </div>
    </div>
  );
  if (wide) return (
    <div style={{ height:'100%', display:'flex', overflow:'hidden' }}>
      <Sidebar active="monitor"/>
      <div style={{ flex:1, display:'flex', flexDirection:'column', overflow:'hidden' }}>
        {topBar}<MonitorContent/>
      </div>
    </div>
  );
  return (
    <div style={{ height:'100%', display:'flex', flexDirection:'column', overflow:'hidden' }}>
      {topBar}<MonitorContent/><TabBar active="monitor"/>
    </div>
  );
}

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   页面：③ 应用列表
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
const APP_LIST = [
  {id:'edge-setting',  type:'系统',s:'running',cpu:'0.3%',mem:'28MB',ver:'v2.0.0'},
  {id:'data-collector',type:'系统',s:'running',cpu:'2.1%',mem:'64MB',ver:'v1.3.2'},
  {id:'modbus-gateway',type:'用户',s:'running',cpu:'1.2%',mem:'45MB',ver:'v1.1.0'},
  {id:'rtsp-proxy',    type:'用户',s:'stopped',cpu:'—',   mem:'—',   ver:'v1.0.3'},
  {id:'ota-agent',     type:'系统',s:'running',cpu:'0.1%',mem:'15MB',ver:'v0.8.1'},
];

function AppListContent() {
  return (
    <div style={{ flex:1, overflowY:'auto', padding:'0 14px' }}>
      <div style={{ textAlign:'center', padding:'10px 0 6px', fontSize:12, color:T.muted }}>↓ 下拉刷新</div>
      {APP_LIST.map(app => {
        const sc = app.s==='running'
          ? {c:T.success, bg:T.successSoft, t:'运行中'}
          : {c:T.danger,  bg:T.dangerSoft,  t:'已停止'};
        return (
          <div key={app.id} style={{ background:T.bgCard, borderRadius:T.r14, padding:'12px 14px',
            marginBottom:10, border:`1px solid ${T.borderSubtle}`,
            borderLeft:`3px solid ${sc.c}`, boxShadow:T.shadowSm }}>
            <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:7 }}>
              <Dot status={app.s}/>
              <span style={{ fontSize:14, fontWeight:700, color:T.text, fontFamily:'monospace', flex:1,
                overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap' }}>{app.id}</span>
              <span style={{ fontSize:11, padding:'2px 7px', borderRadius:6,
                background:app.type==='系统'?'rgba(139,92,246,.1)':T.primarySoft,
                color:app.type==='系统'?T.system:T.primary }}>{app.type}</span>
              <span style={{ fontSize:11, padding:'2px 7px', borderRadius:6, background:sc.bg, color:sc.c }}>{sc.t}</span>
            </div>
            <div style={{ display:'flex', gap:14 }}>
              <span style={{ fontSize:12, color:T.muted }}>CPU {app.cpu}</span>
              <span style={{ fontSize:12, color:T.muted }}>内存 {app.mem}</span>
              <span style={{ fontSize:12, color:T.muted, marginLeft:'auto', fontFamily:'monospace' }}>{app.ver}</span>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function AppsPage({ wide=false }) {
  const topBar = (
    <div style={{ padding:'14px 18px 10px', background:'rgba(255,255,255,0.92)',
      borderBottom:`1px solid ${T.borderSubtle}`, flexShrink:0 }}>
      <div style={{ fontSize:18, fontWeight:700, color:T.text, marginBottom:10 }}>微应用</div>
      {/* 搜索框 */}
      <div style={{ display:'flex', alignItems:'center', gap:8, height:38, borderRadius:T.r12,
        background:T.primarySoft, border:`1px solid ${T.border}`, padding:'0 12px', marginBottom:10 }}>
        <span style={{ fontSize:15, color:T.muted }}>🔍</span>
        <span style={{ fontSize:13, color:T.muted }}>搜索应用名称…</span>
      </div>
      {/* 筛选 Pills */}
      <div style={{ display:'flex', gap:7 }}>
        {['全部  5','运行中  4','已停止  1'].map((f,i)=>(
          <div key={f} style={{ padding:'5px 12px', borderRadius:T.r20, fontSize:12, fontWeight:600, cursor:'pointer',
            background:i===0?T.primary:T.primarySoft, color:i===0?'#fff':T.subText,
            border:`1px solid ${i===0?T.primary:T.border}` }}>{f}</div>
        ))}
      </div>
    </div>
  );
  if (wide) return (
    <div style={{ height:'100%', display:'flex', overflow:'hidden' }}>
      <Sidebar active="apps"/>
      <div style={{ flex:1, display:'flex', flexDirection:'column', overflow:'hidden' }}>
        {topBar}<AppListContent/>
      </div>
    </div>
  );
  return (
    <div style={{ height:'100%', display:'flex', flexDirection:'column', overflow:'hidden' }}>
      {topBar}<AppListContent/><TabBar active="apps"/>
    </div>
  );
}

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   页面：④ 应用详情（只读 + 中文日志）
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
const LOG_ENTRIES = [
  {id:1,lv:'INFO', ts:'10:44:02', msg:'采集周期完成 #1842'},
  {id:2,lv:'INFO', ts:'10:44:00', msg:'MQTT 推送成功'},
  {id:3,lv:'WARN', ts:'10:43:58', msg:'采集延迟 220ms，建议检查网络'},
  {id:4,lv:'INFO', ts:'10:43:55', msg:'心跳发送 OK'},
  {id:5,lv:'ERROR',ts:'10:43:40', msg:'连接超时，已自动重试'},
];

function DetailContent() {
  return (
    <div style={{ flex:1, overflowY:'auto', padding:'0 14px' }}>
      <div style={{ textAlign:'center', padding:'10px 0 6px', fontSize:12, color:T.muted }}>↓ 下拉刷新</div>
      {/* 详情信息卡 */}
      <div style={{ background:T.bgCard, borderRadius:T.r14, overflow:'hidden',
        border:`1px solid ${T.borderSubtle}`, boxShadow:T.shadow, marginBottom:12 }}>
        {[['PID','1002'],['版本','v1.3.2'],['端口','9100'],['启动时间','14天前'],['工作目录','/opt/collector']].map(([k,v],i,arr)=>(
          <div key={k} style={{ display:'flex', justifyContent:'space-between', alignItems:'center',
            padding:'12px 14px', borderBottom:i<arr.length-1?`1px solid ${T.borderSubtle}`:'none' }}>
            <span style={{ fontSize:14, color:T.muted, flexShrink:0 }}>{k}</span>
            <span style={{ fontSize:14, color:T.text, fontFamily:'monospace', fontWeight:600,
              textAlign:'right', marginLeft:12 }}>{v}</span>
          </div>
        ))}
      </div>
      {/* 日志筛选：中文 Pills，非原生 select */}
      <div style={{ display:'flex', alignItems:'center', gap:7, marginBottom:10 }}>
        {['全部','信息','警告','错误'].map((lv,i)=>(
          <div key={lv} style={{ padding:'6px 12px', borderRadius:T.r20, fontSize:12, fontWeight:600, cursor:'pointer',
            background:i===0?T.primary:T.primarySoft, color:i===0?'#fff':T.subText,
            border:`1px solid ${i===0?T.primary:T.border}` }}>{lv}</div>
        ))}
      </div>
      {/* 关键字搜索 */}
      <div style={{ display:'flex', alignItems:'center', gap:8, height:38, borderRadius:T.r12,
        background:T.primarySoft, border:`1px solid ${T.border}`, padding:'0 12px', marginBottom:12 }}>
        <span style={{ fontSize:15, color:T.muted }}>🔍</span>
        <span style={{ fontSize:13, color:T.muted }}>搜索日志内容…</span>
      </div>
      {/* 日志卡片列表 */}
      <div style={{ display:'flex', flexDirection:'column', gap:8, paddingBottom:16 }}>
        {LOG_ENTRIES.map(log=>(
          <div key={log.id} style={{ background:T.bgCard, borderRadius:T.r14, padding:'12px 14px',
            border:`1px solid ${T.borderSubtle}`, boxShadow:T.shadowSm }}>
            <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:7 }}>
              <LvBadge lv={log.lv}/>
              <span style={{ fontSize:12, color:T.muted, fontFamily:'monospace' }}>{log.ts}</span>
            </div>
            <div style={{ fontSize:14, color:T.text, lineHeight:1.6 }}>{log.msg}</div>
          </div>
        ))}
      </div>
    </div>
  );
}

function DetailPage({ wide=false }) {
  const topBar = (
    <div style={{ padding:'12px 16px 0', background:'rgba(255,255,255,0.92)',
      borderBottom:`1px solid ${T.borderSubtle}`, flexShrink:0 }}>
      {/* 返回行 */}
      <div style={{ display:'flex', alignItems:'center', gap:10, marginBottom:10 }}>
        <span style={{ fontSize:26, color:T.subText, lineHeight:1, cursor:'pointer' }}>‹</span>
        <div style={{ flex:1 }}>
          <div style={{ fontSize:15, fontWeight:700, color:T.text, fontFamily:'monospace' }}>data-collector</div>
          <div style={{ fontSize:12, color:T.success, marginTop:2 }}>运行中</div>
        </div>
        {/* 只读状态标签，无操作按钮 */}
        <div style={{ background:T.successSoft, border:`1px solid ${T.successBorder}`,
          borderRadius:T.r8, padding:'5px 12px', fontSize:13, color:T.success, fontWeight:600 }}>运行中</div>
      </div>
      {/* Tab：只有详情 + 日志 */}
      <div style={{ display:'flex', gap:8, paddingBottom:10 }}>
        {['详情','日志'].map((t,i)=>(
          <div key={t} style={{ padding:'6px 18px', borderRadius:T.r20, fontSize:13, fontWeight:600, cursor:'pointer',
            background:i===0?T.primary:T.primarySoft, color:i===0?'#fff':T.subText,
            border:`1px solid ${i===0?T.primary:T.border}` }}>{t}</div>
        ))}
      </div>
    </div>
  );
  if (wide) return (
    <div style={{ height:'100%', display:'flex', overflow:'hidden' }}>
      <Sidebar active="apps"/>
      <div style={{ flex:1, display:'flex', flexDirection:'column', overflow:'hidden' }}>
        {topBar}<DetailContent/>
      </div>
    </div>
  );
  return (
    <div style={{ height:'100%', display:'flex', flexDirection:'column', overflow:'hidden', background:T.bg }}>
      {topBar}<DetailContent/>
    </div>
  );
}

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   页面：⑤ 系统信息（美化布局 + 长文本处理）
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
function InfoRow({ label, value, multiline=false }) {
  if (multiline) return (
    <div style={{ padding:'12px 16px', borderBottom:`1px solid ${T.borderSubtle}` }}>
      <div style={{ fontSize:12, color:T.muted, marginBottom:4 }}>{label}</div>
      <div style={{ fontSize:14, color:T.text, fontFamily:'monospace', fontWeight:600,
        wordBreak:'break-all', lineHeight:1.5 }}>{value}</div>
    </div>
  );
  return (
    <div style={{ display:'flex', justifyContent:'space-between', alignItems:'center',
      padding:'12px 16px', borderBottom:`1px solid ${T.borderSubtle}`, gap:12 }}>
      <span style={{ fontSize:14, color:T.muted, flexShrink:0 }}>{label}</span>
      <span style={{ fontSize:14, color:T.text, fontFamily:'monospace', fontWeight:600,
        textAlign:'right', overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap',
        maxWidth:'60%' }}>{value}</span>
    </div>
  );
}

function AboutContent() {
  return (
    <div style={{ flex:1, overflowY:'auto', padding:'0 14px' }}>
      <div style={{ textAlign:'center', padding:'10px 0 6px', fontSize:12, color:T.muted }}>↓ 下拉刷新</div>
      {/* 节点信息 */}
      <div style={{ background:T.bgCard, borderRadius:T.r14, overflow:'hidden',
        border:`1px solid ${T.borderSubtle}`, boxShadow:T.shadow, marginBottom:12 }}>
        <div style={{ background:T.primarySoft, borderBottom:`1px solid ${T.border}`,
          padding:'8px 16px', display:'flex', alignItems:'center', gap:8 }}>
          <span style={{ fontSize:14 }}>🖥</span>
          <span style={{ fontSize:13, fontWeight:700, color:T.primary, letterSpacing:.3 }}>节点信息</span>
        </div>
        <InfoRow label="主机名" value="edge-node-01"/>
        <InfoRow label="架构" value="amd64"/>
        <div style={{ padding:'12px 16px' }}>
          <div style={{ fontSize:12, color:T.muted, marginBottom:4 }}>系统运行时长</div>
          <div style={{ fontSize:14, color:T.text, fontWeight:600 }}>14天 7时 23分</div>
        </div>
      </div>
      {/* 系统信息 */}
      <div style={{ background:T.bgCard, borderRadius:T.r14, overflow:'hidden',
        border:`1px solid ${T.borderSubtle}`, boxShadow:T.shadow, marginBottom:16 }}>
        <div style={{ background:T.primarySoft, borderBottom:`1px solid ${T.border}`,
          padding:'8px 16px', display:'flex', alignItems:'center', gap:8 }}>
          <span style={{ fontSize:14 }}>⚙</span>
          <span style={{ fontSize:13, fontWeight:700, color:T.primary, letterSpacing:.3 }}>系统信息</span>
        </div>
        {/* 长文本用多行布局 */}
        <div style={{ padding:'12px 16px', borderBottom:`1px solid ${T.borderSubtle}` }}>
          <div style={{ fontSize:12, color:T.muted, marginBottom:4 }}>操作系统</div>
          <div style={{ fontSize:14, color:T.text, fontWeight:600 }}>Ubuntu 22.04 LTS</div>
        </div>
        <div style={{ padding:'12px 16px' }}>
          <div style={{ fontSize:12, color:T.muted, marginBottom:4 }}>内核版本</div>
          <div style={{ fontSize:14, color:T.text, fontFamily:'monospace', fontWeight:600,
            wordBreak:'break-all' }}>5.15.0-91-generic</div>
        </div>
      </div>
    </div>
  );
}

function AboutPage({ wide=false }) {
  const topBar = (
    <div style={{ padding:'14px 18px 10px', background:'rgba(255,255,255,0.92)',
      borderBottom:`1px solid ${T.borderSubtle}`, flexShrink:0 }}>
      <div style={{ fontSize:18, fontWeight:700, color:T.text }}>系统信息</div>
    </div>
  );
  if (wide) return (
    <div style={{ height:'100%', display:'flex', overflow:'hidden' }}>
      <Sidebar active="about"/>
      <div style={{ flex:1, display:'flex', flexDirection:'column', overflow:'hidden' }}>
        {topBar}<AboutContent/>
      </div>
    </div>
  );
  return (
    <div style={{ height:'100%', display:'flex', flexDirection:'column', overflow:'hidden' }}>
      {topBar}<AboutContent/><TabBar active="about"/>
    </div>
  );
}

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   主组件
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
const PAGES = [
  { id:'home',   label:'① 首页',    NarrowComp: ()=><HomePage/>,   WideComp: ()=><HomePage wide/> },
  { id:'monitor',label:'② 监控',    NarrowComp: ()=><MonitorPage/>,WideComp: ()=><MonitorPage wide/> },
  { id:'apps',   label:'③ 应用列表',NarrowComp: ()=><AppsPage/>,  WideComp: ()=><AppsPage wide/> },
  { id:'detail', label:'④ 应用详情',NarrowComp: ()=><DetailPage/>,WideComp: ()=><DetailPage wide/> },
  { id:'about',  label:'⑤ 系统信息',NarrowComp: ()=><AboutPage/>, WideComp: ()=><AboutPage wide/> },
];

export default function EdgeSettingUI() {
  const [cur, setCur] = useState('home');
  const pg = PAGES.find(p => p.id === cur);
  return (
    <div style={{ minHeight:'100vh', background:'#d8e4f8',
      fontFamily:"'DM Sans','Noto Sans SC',sans-serif", padding:'28px 20px 60px' }}>
      <style>{CSS}</style>

      {/* 标题 */}
      <div style={{ textAlign:'center', marginBottom:28 }}>
        <div style={{ display:'inline-flex', alignItems:'center', gap:8, marginBottom:12,
          padding:'5px 18px', borderRadius:T.r20, background:'#fff',
          boxShadow:`0 2px 12px rgba(66,170,245,0.14)`, border:`1px solid ${T.borderSubtle}` }}>
          <div style={{ width:7, height:7, borderRadius:'50%', background:T.primary, animation:'blink 1.5s infinite' }}/>
          <span style={{ fontSize:11, color:T.primary, letterSpacing:2, fontWeight:700 }}>EDGE SETTING · UI v2.0.0</span>
        </div>
        <h1 style={{ fontSize:26, fontWeight:700, color:T.text, letterSpacing:-.3 }}>界面设计稿</h1>
        <p style={{ fontSize:13, color:T.subText, marginTop:5 }}>
          响应式布局 · 窄屏底部 TabBar · 宽屏左侧侧边栏
        </p>
      </div>

      {/* 页面选项卡 */}
      <div style={{ display:'flex', justifyContent:'center', gap:7, marginBottom:30, flexWrap:'wrap' }}>
        {PAGES.map(p => (
          <button key={p.id} onClick={()=>setCur(p.id)} style={{
            padding:'8px 16px', borderRadius:T.r20, border:'none', cursor:'pointer', fontFamily:'inherit',
            background: cur===p.id ? T.primary : '#fff',
            color: cur===p.id ? '#fff' : T.subText,
            fontSize:13, fontWeight:600, transition:'all .2s',
            boxShadow: cur===p.id ? `0 4px 14px rgba(66,170,245,0.38)` : '0 2px 8px rgba(0,0,0,0.06)',
          }}>{p.label}</button>
        ))}
      </div>

      {/* 双屏预览 */}
      <div style={{ display:'flex', justifyContent:'center', alignItems:'flex-start',
        gap:40, marginBottom:50, flexWrap:'wrap' }}>
        <div style={{ display:'flex', flexDirection:'column', alignItems:'center', gap:8 }}>
          <div style={{ fontSize:10, color:T.muted, letterSpacing:3, fontWeight:700 }}>窄屏 · PHONE</div>
          <NarrowFrame label={pg.label}><pg.NarrowComp/></NarrowFrame>
        </div>
        <div style={{ display:'flex', flexDirection:'column', alignItems:'center', gap:8 }}>
          <div style={{ fontSize:10, color:T.muted, letterSpacing:3, fontWeight:700 }}>宽屏 · PAD / DESKTOP</div>
          <WideFrame label={pg.label}><pg.WideComp/></WideFrame>
        </div>
      </div>

      {/* 缩略图总览 */}
      <div style={{ textAlign:'center', marginBottom:12 }}>
        <span style={{ fontSize:10, color:T.subText, letterSpacing:2, fontWeight:700 }}>全部页面缩略图</span>
      </div>
      <div style={{ display:'flex', gap:10, justifyContent:'center', flexWrap:'wrap' }}>
        {PAGES.map(p => (
          <div key={p.id} onClick={()=>setCur(p.id)} style={{ cursor:'pointer', display:'flex', flexDirection:'column', alignItems:'center', gap:6 }}>
            <div style={{ width:100, height:208, background:T.bg, borderRadius:20,
              border: cur===p.id ? `2px solid ${T.primary}` : `1.5px solid ${T.borderSubtle}`,
              overflow:'hidden', position:'relative',
              boxShadow: cur===p.id ? `0 6px 20px rgba(66,170,245,0.28)` : '0 2px 8px rgba(0,0,0,0.06)',
              transform: cur===p.id ? 'scale(1.05)' : 'scale(1)', transition:'all .2s' }}>
              <div style={{ position:'absolute', top:0, left:'50%', transform:'translateX(-50%)',
                width:34, height:8, background:T.bg, borderRadius:'0 0 4px 4px', zIndex:10 }}/>
              <div style={{ transform:'scale(.312)', transformOrigin:'top left',
                width:'320%', height:'312%', position:'absolute', top:0, left:0 }}>
                <div style={{ width:320, height:650, paddingTop:46, overflow:'hidden' }}>
                  <p.NarrowComp/>
                </div>
              </div>
            </div>
            <div style={{ fontSize:10, color: cur===p.id ? T.primary : T.muted, fontWeight: cur===p.id ? 700 : 400 }}>
              {p.label}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
