import reflex as rx

# 

class Header(rx.ComponentState):


    @classmethod
    def get_component(cls, **props):
# def header() -> rx.Component:
        return rx.box(
            rx.heading(
                "Import Virtual Appliance (OVA)", size="6", weight="bold"
            ),
            align_items="center",
            # bg=rx.color("accent", 3),
            # padding="1em",
            # # position="fixed",
            # # top="0px",
            # # z_index="5",
            # width="100%",
            wrap="nowrap",
            # position="fixed",
            justify="between",
            width="100%",
            # top="0",
            align="center",
            left="0",
            # z_index="50",
            padding="1rem",
            background=f"linear-gradient(99deg, {rx.color('red', 2)}, {rx.color('red', 6)}, {rx.color('red', 12)})",
            **props,

    )

header = Header.create