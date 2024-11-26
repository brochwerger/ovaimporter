import reflex as rx

from ovaimporter.templates import template
import ovaimporter.commands as cmd

class OcpData(rx.State):
    url:      str = None
    token:     str = ""  
    namespace: str = ""
    # collected: bool = False

    def handle_submit(self, form_data: dict):
        self.url = form_data['url']
        self.token = form_data['token']
        self.namespace = form_data['namespace']
        result = cmd.ocp_login(self.url, self.token, self.namespace)
        # print(result)
        # result = cmd.ocp_get_pods()
        # print(result)
        return rx.redirect('/')

    def reset_form(self):
        self.url = None
        self.token = ""
        self.namespace = ""
        # self.collected = False

    def isCollected(self):
        if not self.collected:
            return rx.redirect("/ocpinfo")
        
    @rx.var
    def collected(self):
        return self.url is not None

@rx.page(route="/ocpinfo")
@template
def collectOcpDataPage():
    return rx.container(rx.vstack(
        rx.form(
            rx.vstack(
                rx.input(
                    placeholder="OCP server",
                    name="url",
                ),
                rx.input(
                    placeholder="Token",
                    name="token",
                ),
                rx.input(
                    placeholder="Namespace",
                    name="namespace",
                ),
                rx.button("Submit", type="submit"),
            ),
            on_submit=OcpData.handle_submit,
            reset_on_submit=True,
        ),
    ))