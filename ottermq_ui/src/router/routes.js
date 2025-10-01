const routes = [
  {
    path: '/',
    component: () => import('layouts/MainLayout.vue'),
    children: [
      { path: '', component: () => import('pages/IndexPage.vue') },
      { path: 'overview', component: () => import('pages/OverviewPage.vue') },
      { path: 'connections', component: () => import('pages/ConnectionsPage.vue') },
      { path: 'exchanges', component: () => import('pages/ExchangesPage.vue') },
      { path: 'queues', component: () => import('pages/QueuesPage.vue') },
    ],
  },

  // Always leave this as last one,
  // but you can also remove it
  {
    path: '/:catchAll(.*)*',
    component: () => import('pages/ErrorNotFound.vue'),
  },
]

export default routes
