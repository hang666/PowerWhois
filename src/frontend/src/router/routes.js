const routes = [
    {
        path: "/",
        name: "index",
        component: () => import("pages/IndexPage.vue")
    },
    // Always leave this as last one,
    // but you can also remove it
    {
        path: "/:catchAll(.*)*",
        name: "404",
        meta: { title: "404" },
        component: () => import("pages/error/ErrorNotFound.vue")
    }
];

export default routes;
