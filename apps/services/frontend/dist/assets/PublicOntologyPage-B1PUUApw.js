import{ak as C,as as m,a5 as c,r as E,j as l,g as o,d as w,ac as r,ag as t,o as v,S as x,a as T,a6 as k,i as P,al as D,a1 as i,V as $,C as R,F as V,n as A,c as L}from"./index-CFgUTjc4.js";import{g as f}from"./index-BxgMCe-w.js";import{G as N}from"./globe-Du_hRXSe.js";const Q=f`
  query PublicOntologyMetadata($slug: String!) {
    publicOntology(slug: $slug) {
      id
      name
      description
      version
      publishedAt
      classCount
      propertyCount
      individualCount
    }
  }
`,S=f`
  query PublicClassTree($ontologyId: ID!, $maxDepth: Int) {
    classTree(ontologyId: $ontologyId, maxDepth: $maxDepth) {
      id
      label
      children {
        id
        label
        children {
          id
          label
        }
      }
    }
  }
`,B=f`
  query PublicProperties($ontologyId: ID!) {
    properties(ontologyId: $ontologyId, perPage: 100) {
      items {
        id
        label
        propertyType
        xsdType
      }
    }
  }
`;function U(I){const y=c(!0),b=c(null),n=c(null),p=c([]),u=c([]),g=c(null),d=c(""),_=C(Q,{slug:I});m(_.result,s=>{if(!(s!=null&&s.publicOntology))return;const e=s.publicOntology;n.value={id:e.id,name:e.name,description:e.description,version:e.version,publishedAt:e.publishedAt,classCount:e.classCount,propertyCount:e.propertyCount,individualCount:e.individualCount},d.value=e.id,y.value=!1}),m(_.error,s=>{s&&(b.value=s.message,y.value=!1)});const h=C(S,()=>({ontologyId:d.value,maxDepth:3}));m(h.result,s=>{s!=null&&s.classTree&&(p.value=s.classTree)});const O=C(B,()=>({ontologyId:d.value}));m(O.result,s=>{var e;(e=s==null?void 0:s.properties)!=null&&e.items&&(u.value=s.properties.items)});function a(s){g.value=s}return{metadata:n,classTree:p,properties:u,loading:y,error:b,selectedClassId:g,selectClass:a}}const z={class:"public-page",role:"main","aria-label":"Public Ontology View"},Y={class:"public-header"},q={class:"public-center"},F={key:0,class:"public-loading"},G={key:1,class:"public-error"},M={key:2,class:"public-main"},j={class:"public-class card-side"},H={class:"public-tools"},J=["onClick"],K={key:0,class:"public-tree-row muted"},W={class:"public-graph card-side"},X={key:0,class:"public-empty"},Z={class:"public-props card-side"},ss={class:"muted"},es={key:0,class:"muted prop-row"},os={class:"public-banner"},ts=E({__name:"PublicOntologyPage",setup(I){const b=D().params.id||"default",{metadata:n,classTree:p,properties:u,loading:g,error:d,selectedClassId:_,selectClass:h}=U(b);return(O,a)=>{var s;return i(),l("div",z,[o("header",Y,[a[0]||(a[0]=o("div",{class:"public-brand"},[o("img",{src:w,alt:"VEDO Core",class:"public-brand-logo"}),o("h1",{class:"public-brand-text",role:"heading"},"Public Ontology")],-1)),o("span",q,r(((s=t(n))==null?void 0:s.name)??"Public Ontology"),1),a[1]||(a[1]=o("span",{class:"public-readonly"},"Read-only",-1))]),t(g)?(i(),l("div",F,"Loading ontology...")):t(d)?(i(),l("div",G,r(t(d)),1)):(i(),l("div",M,[o("aside",j,[o("div",H,[v(t(x),{size:14,class:"muted"}),a[2]||(a[2]=o("div",{class:"panel-input"},"Filter classes...",-1))]),(i(!0),l(T,null,k(t(p),e=>(i(),l("div",{key:e.id,class:$(["public-tree-row",{"public-tree-row--active":t(_)===e.id}]),onClick:as=>t(h)(e.id)},[v(t(R),{size:12}),v(t(V),{size:14}),A(" "+r(e.label),1)],10,J))),128)),t(p).length===0?(i(),l("div",K,"No classes found")):P("",!0)]),o("section",W,[a[3]||(a[3]=o("div",{class:"graph-head"},[o("span",{class:"col-ind"},"Individual"),o("span",{class:"col-prop"},"Property"),o("span",{class:"col-val"},"Value")],-1)),t(n)?(i(),l("div",X,"Published ontology with "+r(t(n).classCount)+" classes, "+r(t(n).propertyCount)+" properties, "+r(t(n).individualCount)+" individuals",1)):P("",!0)]),o("aside",Z,[a[4]||(a[4]=o("h2",{class:"prop-title"},"Properties",-1)),(i(!0),l(T,null,k(t(u),e=>(i(),l("div",{key:e.id,class:"prop-row"},[o("span",null,r(e.label),1),o("span",ss,r(e.propertyType),1)]))),128)),t(u).length===0?(i(),l("div",es,"No properties")):P("",!0)])])),o("footer",os,[v(t(N),{size:14,class:"muted"}),a[5]||(a[5]=o("span",null,"You are viewing a published snapshot. Edits are disabled.",-1))])])}}}),rs=L(ts,[["__scopeId","data-v-5866a881"]]);export{rs as default};
